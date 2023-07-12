package job

import "strings"

type Broker struct {
	pubCh   chan string
	subCh   chan chan string
	unsubCh chan chan string
	stop    chan struct{}
}

func NewBroker() *Broker {
	return &Broker{
		pubCh:   make(chan string, 1),
		subCh:   make(chan chan string, 1),
		unsubCh: make(chan chan string, 1),
		stop:    make(chan struct{}, 1),
	}
}

func (b *Broker) Start() {
	go func() {
		subs := map[chan string]struct{}{}

		for {
			select {
			case <-b.stop:
				for msgCh := range subs {
					close(msgCh)
				}
				return
			case msgCh := <-b.subCh:
				subs[msgCh] = struct{}{}
			case msgCh := <-b.unsubCh:
				delete(subs, msgCh)
			case msg := <-b.pubCh:
				for msgCh := range subs {
					select {
					case msgCh <- msg:
					default: // buffer full
						<-msgCh
						msgCh <- msg
					}
				}
			}
		}
	}()
}

func (b *Broker) Publish(msg string) {
	b.pubCh <- msg
}

func (b *Broker) Subscribe() chan string {
	msgCh := make(chan string, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan string) {
	b.unsubCh <- msgCh
}

func (b *Broker) Stop() {
	close(b.stop)
}

// Write is a wrapper around Publish to implement the io.Writer interface.
func (b *Broker) Write(p []byte) (n int, err error) {
	b.pubCh <- strings.TrimSuffix(string(p), "\n")
	return len(p), nil
}
