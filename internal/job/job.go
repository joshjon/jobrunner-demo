package job

import (
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

type State int

const (
	StateUnspecified State = iota
	StateQueued
	StateRunning
	StateCompleted
)

type Command struct {
	Cmd  string
	Args []string
}

type Job struct {
	ID     string
	state  State
	mu     sync.RWMutex
	cmd    *exec.Cmd
	broker *Broker
}

func NewJob(command Command) (*Job, error) {
	broker := NewBroker()
	broker.Start()
	cmd := exec.Command(command.Cmd, command.Args...)
	cmd.Stdout = broker
	cmd.Stderr = broker
	return &Job{
		ID:     uuid.New().String(),
		state:  StateQueued,
		cmd:    cmd,
		broker: broker,
	}, nil
}

func (j *Job) Run() error {
	j.mu.RLock()
	j.state = StateRunning
	j.mu.RLock()
	defer func() {
		j.state = StateCompleted
		j.broker.Stop()
	}()
	if err := j.cmd.Start(); err != nil {
		return err
	}
	if err := j.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}
	return nil
}

func (j *Job) GetState() State {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state
}

func (j *Job) updateState(state State) {
	j.mu.Lock()
	j.state = state
	j.mu.Lock()
}
