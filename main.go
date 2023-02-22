// package gvents is a very simple go event (pubsub) module.
// gvents is concurrency-safe mostly employing sync.Map for keeping track of events and subscriptions.
// gvents uses a system of uuids and subscriber-name to allow easier debugging, retrieval and safe unsubscribing.
// gvents also comes with some convenience functions like subscriber listing and counting mainly nice for debugging.
// gvents calls all event-handlers "in parallel" using go routines - this should be efficient, but handle with care ;)
// use as you like - submit issues here: https://github.com/itsatony/gvents
package gvents

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

var ErrFailedToSubscribe error = errors.New("failed to subscribe")

// Event is the event struct used by gvents
type Event struct {
	Name   string
	Data   any
	Cancel bool
} // @name Event

// PubSub is the main struct for gvents
type PubSub struct {
	subscribers sync.Map
	states      sync.Map // string -> any
} // @name PubSub

// NewPubSub creates a new PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: sync.Map{},
		states:      sync.Map{},
	}
}

// Subscribe subscribes a handler to an event
func (p *PubSub) Subscribe(eventName string, handlerName string, handler func(*Event)) (handlerId string, err error) {
	err = nil
	handlerId = uuid.New().String() + ":" + handlerName
	subs := &sync.Map{}
	subsAny, loaded := p.subscribers.LoadOrStore(eventName, subs)
	if loaded {
		var ok bool
		subs, ok = subsAny.(*sync.Map)
		if !ok {
			subs = &sync.Map{}
		}
	}
	subs.Store(handlerId, handler)
	p.subscribers.Store(eventName, subs)
	// DEBUG
	// fmt.Printf("===> SUBSCRIBE for(%s) by(%s). full sub list: \n%s\n\n", eventName, handlerId, p.SubList(eventName))
	return handlerId, err
}

// Unsubscribe unsubscribes a handler from an event
func (p *PubSub) Unsubscribe(eventName string, id string) (foundEvent bool, foundAndDeletedHandler bool) {
	foundAndDeletedHandler = false // // DEBUG
	// fmt.Printf("===> UN-SUBSCRIBE: \n%s\n", p.SubList(eventName))
	subsAny, ok := p.subscribers.Load(eventName)
	if ok {
		foundEvent = true // // DEBUG
		// fmt.Printf("===> UN-SUBSCRIBE: event (%s) found\n", eventName)
		subs, ok := subsAny.(*sync.Map)
		if ok {
			_, foundAndDeletedHandler = subs.LoadAndDelete(id)
			// } else {			// 	// DEBUG
			// 	fmt.Println("===> UN-SUBSCRIBE: casting failed")
		}
	} else {
		foundEvent = false
	}
	return foundEvent, foundAndDeletedHandler
}

// PublishAsState publishes an event and set the eventname + data as state
func (p *PubSub) PublishAsState(thisEvent *Event) (eventFound bool, handlerIds []string) {
	p.states.Store(thisEvent.Name, thisEvent.Data)
	return p.Publish(thisEvent)
}

// Publish publishes an event
func (p *PubSub) Publish(thisEvent *Event) (eventFound bool, handlerIds []string) {
	eventFound = false
	handlerIds = []string{}
	subsAny, ok := p.subscribers.Load(thisEvent.Name)
	if !ok {
		return eventFound, handlerIds
	}
	eventFound = true
	subs, ok := subsAny.(*sync.Map)
	if !ok {
		// fmt.Errorf("subs are not a sync map!? %t\n", ok)
		return eventFound, handlerIds
	}
	subs.Range(func(key any, value any) bool {
		id, ok := key.(string)
		if !ok {
			// something's VERY wrong here.. should probably delete this value or panic?
			// fmt.Errorf("sub id is not a string!? %s\n", key)
			return true
		}
		handlerFunc, ok := value.(func(*Event))
		if !ok {
			// something's VERY wrong here.. should probably delete this value or panic?
			// fmt.Errorf("handler is not a valid func? %s\n", value)
			return true
		}
		handlerIds = append(handlerIds, id)
		go handlerFunc(thisEvent)
		return true
	})
	return eventFound, handlerIds
}

// CancelEvent cancels an event
// since event handling is done in parallel, this is not a guarantee that the event will not be handled
// in fact, it is very likely that the event will be handled before the cancellation hits
func (p *PubSub) CancelEvent(eventName string) (eventFound bool, handlerIds []string) {
	ev := &Event{Name: eventName, Cancel: true, Data: nil}
	return p.Publish(ev)
}

// EventExists checks if an event exists by checking if it has subscribers
func (p *PubSub) EventExists(eventName string) bool {
	_, ok := p.subscribers.Load(eventName)
	return ok
}

// HasSubscribedTo checks if a subscriber has subscribed to an event
func (p *PubSub) HasSubscribedTo(eventName string, subscriberId string) bool {
	subsAny, ok := p.subscribers.Load(eventName)
	if !ok {
		return false
	}
	subs, ok := subsAny.(*sync.Map)
	if ok {
		_, foundHandler := subs.Load(subscriberId)
		return foundHandler
	}
	return false
}

// SubCount returns the number of subscribers for an event
func (p *PubSub) SubCount(eventName string) (subCount int) {
	subCount = -1
	subsAny, ok := p.subscribers.Load(eventName)
	if !ok {
		return subCount
	}
	subs, ok := subsAny.(*sync.Map)
	if !ok {
		return subCount
	}
	subCount = 0
	subs.Range(func(key any, value any) bool {
		subCount += 1
		return true
	})
	return subCount
}

// SubList returns a list of subscribers for an event
func (p *PubSub) SubList(eventName string) (sublist []string) {
	sublist = []string{}
	subsAny, ok := p.subscribers.Load(eventName)
	if !ok {
		return sublist
	}
	subs, ok := subsAny.(*sync.Map)
	if !ok {
		return sublist
	}
	subs.Range(func(key any, value any) bool {
		id, ok := key.(string)
		if ok {
			sublist = append(sublist, id)
		}
		return true
	})
	return sublist
}

// SetState sets a state
func (p *PubSub) SetState(key string, value any) {
	p.states.Store(key, value)
}

// GetState gets a state
func (p *PubSub) GetState(key string) (value any, found bool) {
	return p.states.Load(key)
}

// DeleteState deletes a state
func (p *PubSub) DeleteState(key string) {
	p.states.Delete(key)
}

// ClearStates deletes all states
func (p *PubSub) ClearStates() {
	p.states.Range(func(key any, value any) bool {
		p.states.Delete(key)
		return true
	})
}

// SetStates sets multiple states
func (p *PubSub) SetStates(states map[string]any) {
	for key, value := range states {
		p.states.Store(key, value)
	}
}
