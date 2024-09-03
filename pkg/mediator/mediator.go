package mediator

import (
	"fmt"
	"sync"
)

const (
	workers      = 5
	jobsChanSize = 10000
)

type Dispatcher struct {
	jobs        chan Job
	events      map[EventName]Listener
	afterEvents map[EventName]EventName
	mutex       *sync.Mutex
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{
		jobs:        make(chan Job, jobsChanSize),
		events:      make(map[EventName]Listener),
		afterEvents: make(map[EventName]EventName),
		mutex:       &sync.Mutex{},
	}
	for i := 0; i < workers; i++ {
		go d.consume()
	}
	return d
}

func (d *Dispatcher) GetEvent(name EventName) (Listener, bool) {
	d.mutex.Lock()
	result, ok := d.events[name]
	d.mutex.Unlock()
	return result, ok
}

func (d *Dispatcher) SetEvent(name EventName, listener Listener) {
	d.mutex.Lock()
	d.events[name] = listener
	d.mutex.Unlock()
}

func (d *Dispatcher) GetAfterEvent(name EventName) (EventName, bool) {
	d.mutex.Lock()
	result, ok := d.afterEvents[name]
	d.mutex.Unlock()
	return result, ok
}

func (d *Dispatcher) SetAfterEvent(name EventName, listener EventName) {
	d.mutex.Lock()
	d.afterEvents[name] = listener
	d.mutex.Unlock()
}

func (d *Dispatcher) Register(listener Listener, names ...EventName) error {
	for _, name := range names {
		if _, ok := d.GetEvent(name); ok {
			return fmt.Errorf("the '%s' event is already registered", name)
		}
		d.SetEvent(name, listener)
	}

	return nil
}

func (d *Dispatcher) Dispatch(name EventName, event interface{}) error {
	if _, ok := d.GetEvent(name); !ok {
		return fmt.Errorf("the '%s' event is not registered", name)
	}

	d.jobs <- Job{EventName: name, EventType: event}

	if _, ok := d.afterEvents[name]; ok {
		d.jobs <- Job{EventName: d.afterEvents[name], EventType: event}
	}

	return nil
}

func (d *Dispatcher) consume() {
	var listener Listener
	for job := range d.jobs {
		listener, _ = d.GetEvent(job.EventName)
		//go listener.Listen(job.EventName, job.EventType)
		listener.Push(job.EventName, job.EventType) //TODO: or go push? Add check limits
	}
}
