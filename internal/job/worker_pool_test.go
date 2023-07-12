package job

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	workerCount := 2
	numJobs := workerCount + 1
	numSubs := 3

	p := NewWorkerPool()
	errs := p.Start(workerCount)

	var wg sync.WaitGroup

	for i := 0; i < numJobs; i++ {
		cmd := echoLoop(3, 2.0, i)
		cmd = Command{
			Cmd:  "ping",
			Args: []string{"-c", "2", "google.com"},
		}
		id, err := p.QueueJob(cmd)
		require.NoError(t, err)
		t.Logf("job %s queued", id)

		wg.Add(numSubs)
		go func() {
			for j := 0; j < numSubs; j++ {
				go func(i int) {
					sub, err := p.Subscribe(id)
					require.NoError(t, err)

					defer wg.Done()
					for {
						select {
						case err = <-errs:
							require.NoError(t, err)
							return
						case msg, ok := <-sub:
							if !ok {
								return
							}
							t.Logf("subcriber %d: %s", i, msg)
						}
					}
				}(j)
			}
		}()
	}

	wg.Wait()
}

func echoLoop(iterations int, delay float64, id int) Command {
	return Command{
		Cmd: "bash",
		Args: []string{"-c", fmt.Sprintf("for i in {1..%d}; do echo job %d log ${i}; sleep %f; done",
			iterations, id, delay)},
	}
}
