package job

import (
	"context"

	servicev1 "github.com/joshjon/jobrunner-demo/gen/rpc/service/v1"
)

var _ servicev1.ServiceServer = (*Server)(nil)

type Pool interface {
	Start(workerCount int) chan error
	QueueJob(cmd Command) (string, error)
	StopJob(id string) error
	Subscribe(id string) (chan string, error)
	Unsubscribe(id string, sub chan string) error
}

type Server struct {
	pool Pool
}

func NewServer(workerCount int) (*Server, chan error) {
	pool := NewWorkerPool()
	errs := pool.Start(workerCount)
	return &Server{
		pool: pool,
	}, errs
}

func (s *Server) QueueJob(ctx context.Context, req *servicev1.QueueJobRequest) (*servicev1.QueueJobResponse, error) {
	id, err := s.pool.QueueJob(Command{
		Cmd:  req.Command.Cmd,
		Args: req.Command.Args,
	})
	if err != nil {
		return nil, err
	}
	return &servicev1.QueueJobResponse{
		JobId: id,
	}, nil
}

func (s *Server) StopJob(ctx context.Context, req *servicev1.StopJobRequest) (*servicev1.StopJobResponse, error) {
	return &servicev1.StopJobResponse{}, s.pool.StopJob(req.JobId)
}

func (s *Server) Subscribe(req *servicev1.SubscribeRequest, stream servicev1.Service_SubscribeServer) error {
	sub, err := s.pool.Subscribe(req.JobId)
	if err != nil {
		return err
	}
	defer s.pool.Unsubscribe(req.JobId, sub)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg, ok := <-sub:
			if !ok {
				return nil
			}
			if err = stream.Send(&servicev1.SubscribeResponse{
				Message: msg,
			}); err != nil {
				return err
			}
		}
	}
}
