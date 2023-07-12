package job

import (
	"errors"
	"fmt"
	"sync"
)

const bufferSize = 100

var NotFoundError = errors.New("job not found")
var SubscribeJobCompletedError = errors.New("unable to subscribe: job completed")

type WorkerPool struct {
	jobsCh      chan *Job
	currentJobs *sync.Map
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		jobsCh:      make(chan *Job, bufferSize),
		currentJobs: new(sync.Map),
	}
}

func (p *WorkerPool) Start(workerCount int) chan error {
	errs := make(chan error, bufferSize)

	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			for j := range p.jobsCh {
				p.currentJobs.Store(j.ID, j)

				if err := j.Run(); err != nil {
					errs <- fmt.Errorf("worker %d job %s failed: %w", workerID, j.ID, err)
				}
			}
		}(i)
	}

	return errs
}

func (p *WorkerPool) QueueJob(cmd Command) (string, error) {
	job, err := NewJob(cmd)
	if err != nil {
		return "", err
	}
	p.currentJobs.Store(job.ID, job)
	p.jobsCh <- job
	return job.ID, nil
}

func (p *WorkerPool) StopJob(id string) error {
	val, ok := p.currentJobs.Load(id)
	if !ok {
		return NotFoundError
	}
	job := val.(*Job)
	if job.GetState() != StateRunning {
		return nil
	}
	return job.cmd.Process.Kill()
}

func (p *WorkerPool) Subscribe(id string) (chan string, error) {
	val, ok := p.currentJobs.Load(id)
	if !ok {
		return nil, NotFoundError
	}
	job := val.(*Job)
	state := job.GetState()
	switch state {
	case StateQueued:
		sub := job.broker.Subscribe()
		sub <- "info: waiting for available worker, job will start shortly..."
		return sub, nil
	case StateRunning:
		return job.broker.Subscribe(), nil
	case StateCompleted:
		return nil, SubscribeJobCompletedError
	default:
		return nil, errors.New("job state unknown")
	}
}

func (p *WorkerPool) Unsubscribe(id string, sub chan string) error {
	val, ok := p.currentJobs.Load(id)
	if !ok {
		return NotFoundError
	}
	val.(*Job).broker.Unsubscribe(sub)
	return nil
}
