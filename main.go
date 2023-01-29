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

type Event struct {
	Name   string
	Data   any
	Cancel bool
} // @name Event

type PubSub struct {
	subscribers sync.Map
} // @name PubSub

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: sync.Map{},
	}
}

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

func (p *PubSub) CancelEvent(eventName string) (eventFound bool, handlerIds []string) {
	ev := &Event{Name: eventName, Cancel: true, Data: nil}
	return p.Publish(ev)
}

func (p *PubSub) EventExists(eventName string) bool {
	_, ok := p.subscribers.Load(eventName)
	return ok
}

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
