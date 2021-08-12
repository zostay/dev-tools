package future

import (
	"sync"
)

type StartedPromise struct {
	ready *sync.WaitGroup
	val   interface{}
	err   error
}

type DeferredPromise struct {
	StartedPromise
}

type Promise interface {
	Get() (interface{}, error)
	Then(Followup) Promise
}

type Actor func() (interface{}, error)
type Followup func(interface{}, error) (interface{}, error)

func Start(start Actor) *StartedPromise {
	p := &StartedPromise{ready: new(sync.WaitGroup)}
	p.ready.Add(1)
	go func() {
		defer p.ready.Done()
		p.val, p.err = start()
	}()

	return p
}

func Deferred() *DeferredPromise {
	p := &DeferredPromise{
		StartedPromise: StartedPromise{ready: new(sync.WaitGroup)},
	}
	p.ready.Add(1)
	return p
}

func (p *DeferredPromise) When(q Promise) {
	go func() {
		defer p.ready.Done()
		p.val, p.err = q.Get()
	}()
}

func (p *DeferredPromise) Keep(val interface{}, err error) {
	defer p.ready.Done()
	p.val = val
	p.err = err
}

func (p *StartedPromise) Get() (interface{}, error) {
	p.ready.Wait()
	return p.val, p.err
}

func (p *StartedPromise) Then(then Followup) Promise {
	q := &StartedPromise{ready: new(sync.WaitGroup)}
	q.ready.Add(1)

	go func() {
		defer q.ready.Done()
		val, err := p.Get()
		q.val, q.err = then(val, err)
	}()

	return q
}
