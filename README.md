# gvents

gvents is a very simple go event (pubsub) module.
Its first iteration was written in a (pretty long) dialogue with ChatGPT (as a demo and test). In the end, the generated code was unusable and so i refactored 95% of it to the current state.
gvents is concurrency-safe mostly employing sync.Map for keeping track of events and subscriptions.
gvents uses a system of uuids and subscriber-name to allow easier debugging, retrieval and safe unsubscribing.
gvents also comes with some convenience functions like subscriber listing and counting mainly nice for debugging.
gvents calls all event-handlers "in parallel" using go routines - this should be efficient, but handle with care ;)

## Methods of the publisher

* Subscribe(eventName string, handlerName string, handler func(*Event)) (handlerId string, err error)
* Unsubscribe(eventName string, id string)  (foundEvent bool, foundAndDeletedHandler bool)
* Publish(thisEvent *Event) (eventFound bool, handlerIds []string)
* CancelEvent(eventName string) (eventFound bool, handlerIds []string)
* EventExists(eventName string) bool
* HasSubscribedTo(eventName string, subscriberId string) bool
* SubCount(eventName string) (subCount int)
* SubList(eventName string) (sublist []string)

## the major method are tested for base functionality

## example use

````go
// subscribes, checks its subscription and publishes using a scoped var to demo go-routine concurrency
// check the tests for more example uses
func SubPub() {
 var eventName1 string = "testEvent1"
 var handlerName1 string = "handlerName1"
 ps := NewPubSub()
 var callbackInvoked bool = false
 handler := func(e *Event) {
  callbackInvoked = true
  fmt.Println("YAY, this super callback for eventName1 was invoked")
 }
 subId, err := ps.Subscribe(eventName1, handlerName1, handler)
 if err != nil {
  fmt.Println(err)
 }
 ok := ps.HasSubscribedTo(eventName1, subId)
 if !ok {
  fmt.Println("subscription failed")
 }
 found, ids := ps.Publish(&Event{Name: eventName1})
 fmt.Printf("found=(%t)\ncalled handlers:(%s)\n", found, ids)
 // this sleep is bad form of course, but should be fine here ... 
 // it is used here to ensure that the handler, which is called in a go-routine and modifies a scoped var, completes before we evaluate the result.
 time.Sleep(time.Millisecond * 300)
 if !callbackInvoked {
  fmt.Println("Callback was not invoked when event was emitted")
 }
}
````

## VERSIONS

* v0.1.0 initial version of the gvents package.

## TODO

* nothing so far - feel free to suggest stuff
