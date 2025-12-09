package hub

import (
	"sync"

	"deepapp_golang_grpc_hub/internal/proto"
)

type Dispatcher struct {
	queue chan *proto.Message
	wg    sync.WaitGroup
}

func NewDispatcher(router *Router) *Dispatcher {
	d := &Dispatcher{
		queue: make(chan *proto.Message, 100),
	}
	d.start(router)
	return d
}

func (d *Dispatcher) start(router *Router) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for msg := range d.queue {
			router.Route(msg)
		}
	}()
}

func (d *Dispatcher) Dispatch(msg *proto.Message) {
	d.queue <- msg
}

func (d *Dispatcher) Stop() {
	close(d.queue)
	d.wg.Wait()
}