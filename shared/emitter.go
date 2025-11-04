package shared

type EventEmitter[T any] struct {
	events map[string][]func(T)
}

func NewEventEmitter[T any]() *EventEmitter[T] {
	return &EventEmitter[T]{
		events: map[string][]func(T){},
	}
}

/*
*
Listen on something.
*/
func (e *EventEmitter[T]) On(eventName string, handler func(T)) {
	if e.events[eventName] == nil {
		e.events[eventName] = []func(T){}
	}
	list := e.events[eventName]
	list = append(list, handler)
	e.events[eventName] = list
}

/*
*
Event event.
*/
func (e *EventEmitter[T]) Emit(eventName string, data T) {
	list := e.events[eventName]
	for _, handler := range list {
		go handler(data)
	}
}
