package mediator

type Listener interface {
	Listen(eventName EventName, event interface{})
	Push(eventName EventName, event interface{})
}

type Job struct {
	EventName EventName
	EventType interface{}
}

type EventName string
