package gvents

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Run all the tests
	code := m.Run()
	// Exit with the returned code
	os.Exit(code)
}

func TestSubscribe(t *testing.T) {
	var eventName1 string = "testEvent1"
	var handlerName1 string = "handlerName1"
	ps := NewPubSub()
	handler := func(e *Event) {}
	subscriberId, err := ps.Subscribe(eventName1, handlerName1, handler)
	if err != nil {
		t.Error(err)
	}
	isSubbed := ps.HasSubscribedTo(eventName1, subscriberId)
	if !isSubbed {
		t.Error("no matching subscriber found")
	}
}

func TestUnsubscribe(t *testing.T) {
	var eventName1 string = "testEvent1"
	var handlerName1 string = "handlerName1"
	ps := NewPubSub()
	handler := func(e *Event) {}
	subscriberId, err := ps.Subscribe(eventName1, handlerName1, handler)
	if err != nil {
		t.Error(err)
	}
	f1, f2 := ps.Unsubscribe(eventName1, subscriberId)
	if !f1 {
		t.Error("event not found")
	}
	if !f2 {
		t.Error("subscriber not found")
	}
}

func TestPublish(t *testing.T) {
	var eventName1 string = "testEvent1"
	var handlerName1 string = "handlerName1"
	ps := NewPubSub()
	var callbackInvoked bool = false
	handler := func(e *Event) {
		callbackInvoked = true
	}
	subId, err := ps.Subscribe(eventName1, handlerName1, handler)
	if err != nil {
		t.Error(err)
	}
	ok := ps.HasSubscribedTo(eventName1, subId)
	if !ok {
		t.Error("subscription failed")
	}
	found, ids := ps.Publish(&Event{Name: eventName1})
	fmt.Printf("found=(%t)\nhandlers=(%s)\n", found, ids)
	// this is bad form, but should be fine
	time.Sleep(time.Millisecond * 300)
	if !callbackInvoked {
		t.Error("Callback was not invoked when event was emitted")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	var eventName1 string = "testEvent1"
	var eventName2 string = "testEvent2"
	var subInt1 int = 0
	var subInt2 int = 0
	ps := NewPubSub()
	callback1 := func(event *Event) { subInt1 += 1 }
	callback2 := func(event *Event) { subInt1 += 10 }
	callback3 := func(event *Event) { subInt1 += 100 }
	callback4 := func(event *Event) { subInt2 += 1000 }
	callback5 := func(event *Event) { subInt2 += 10000 }
	ps.Subscribe(eventName1, "sub1", callback1)
	ps.Subscribe(eventName1, "sub2", callback2)
	ps.Subscribe(eventName1, "sub3", callback3)
	ps.Subscribe(eventName2, "sub1", callback4)
	ps.Subscribe(eventName2, "sub3", callback5)
	count := ps.SubCount(eventName1)
	if count != 3 {
		t.Errorf("Expected 3 subscribers for (%s), got (%d)", eventName1, count)
	}
	count = ps.SubCount(eventName2)
	if count != 2 {
		t.Errorf("Expected 2 subscribers for (%s), got (%d)", eventName2, count)
	}
	ps.Publish(&Event{Name: eventName1, Data: nil})
	ps.Publish(&Event{Name: eventName2, Data: nil})
	// this is bad form, but should be fine
	time.Sleep(time.Millisecond * 100)
	if subInt1 != 111 {
		t.Errorf("Expected 111 as summed value for all handlers for (%s) but got (%d)", eventName1, subInt1)
	}
	if subInt2 != 11000 {
		t.Errorf("Expected 101 as summed value for all handlers for (%s) but got (%d)", eventName1, subInt1)
	}
}

// func TestUnsubscribeNonExisting(t *testing.T) {
// 	// Try to unsubscribe a non-existing subscriber
// 	Unsubscribe("event1", "non-existing-subscriber")

// 	// Check if there are any issues
// 	if len(subscribers.Load().(map[string][]subscriber)) != 0 {
// 		t.Error("Unsubscribing a non-existing subscriber caused issues")
// 	}
// }

func TestSetState(t *testing.T) {
	var stateName1 string = "testState1"
	var stateValue1 bool = true
	ps := NewPubSub()
	ps.SetState(stateName1, stateValue1)
	state, found := ps.GetState(stateName1)
	if !found {
		t.Error("state not found")
	}
	if state != stateValue1 {
		t.Error("state not set")
	}
}

func TestStateEquals(t *testing.T) {
	var stateName1 string = "testState1"
	var stateValue1 bool = true
	ps := NewPubSub()
	ps.SetState(stateName1, stateValue1)
	ok := ps.StateEquals(stateName1, stateValue1)
	if !ok {
		t.Error("state not equal")
	}
}

func TestStateNotEquals(t *testing.T) {
	var stateName1 string = "testState1"
	var stateValue1 bool = true
	ps := NewPubSub()
	ps.SetState(stateName1, stateValue1)
	ok := ps.StateEquals(stateName1, !stateValue1)
	if ok {
		t.Error("state equal")
	}
}

func TestStateNotExists(t *testing.T) {
	var stateName1 string = "testState1"
	ps := NewPubSub()
	_, found := ps.GetState(stateName1)
	if found {
		t.Error("state found")
	}
}

func TestHasState(t *testing.T) {
	var stateName1 string = "testState1"
	var stateValue1 bool = true
	ps := NewPubSub()
	ps.SetState(stateName1, stateValue1)
	ok := ps.HasState(stateName1)
	if !ok {
		t.Error("state not found")
	}
}

func TestHasNoState(t *testing.T) {
	var stateName1 string = "testState1"
	ps := NewPubSub()
	ok := ps.HasState(stateName1)
	if ok {
		t.Error("state found")
	}
}

func TestSetStates(t *testing.T) {
	var stateName1 string = "testState1"
	var stateValue1 bool = true
	var stateName2 string = "testState2"
	var stateValue2 bool = false
	ps := NewPubSub()
	ps.SetStates(map[string]interface{}{
		stateName1: stateValue1,
		stateName2: stateValue2,
	})
	state, found := ps.GetState(stateName1)
	if !found {
		t.Error("state not found")
	}
	if state != stateValue1 {
		t.Error("state not set")
	}
	state, found = ps.GetState(stateName2)
	if !found {
		t.Error("state not found")
	}
	if state != stateValue2 {
		t.Error("state not set")
	}
}
