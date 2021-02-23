
// Global handling of: transitions, state locking, errors, retries, 

```go
type ThingsYouCanDo struct {
	Log(msg string, a ...interface{})
	Throw(code, msg string, a ...interface{})
	Set(key string, interface{}) error
	JQ(command string) (interface{}, error)
	Transform(command string) error
	Save(savedata []byte) error
	Yield() error
	Crash(stateLogic StateLogic, msg string, a ...interface{})
}

type DelayState struct {

}

func (s *DelayState) Begin(ctx context.Context, state State) {
	state.QueueSleep(blah)

	s.engine.Sleep(blah)
}

type ActionResult struct {
	Output interface{}
	Error struct {
		Code string 
		Message string
	}
}

type StateLogic interface {
	FailTimeout() time.Duration
	Cleanup() // Does this work?
	Run(ctx context.Context, state State, savedata, wakedata []byte)
}
```


```go
type CallbackStateLogic struct {

}

func (sl *CallbackStateLogic) FailTimeout() time.Duration {
	return sl.state.timeout + time.Minute*5
}

func (sl *CallbackStateLogic) Run(ctx context.Context, state State, savedata, wakedata []byte) {

	var err error
	step := string(savedata)

	if step == "" {

		err = sl.reaper.Timeout(sl)
		if err != nil {
			state.Crash(sl, "failed to initialize reaper: %v", err)
			return
		}

		err = sl.eventsEngine.RegisterInterest(sl.InstanceID, sl.namespace, sl.state.Event.Type)
		if err != nil {
			state.Crash(sl, "failed to register interest in event: %v", err)
			return
		}

		err = state.Save([]byte("waitingOnAction"))
		if err != nil {
			state.Crash(sl, "failed to save state information: %v", err)
			return
		}

		err = state.Yield()
		if err != nil {
			state.Crash(sl, "failed to yield state lock: %v", err)
			return
		}

		err = sl.actionsEngine.Queue(sl.state.Action)
		if err != nil {
			state.Crash(sl, "failed to queue action: %v", err)
			return
		}

		return
	}

	if step == "waitingOnAction" {

		actionResults := new(ActionResults)

		err = json.Unmarshal(wakedata, actionResults)
		if err != nil {
			state.Crash(sl, "failed to unmarshal action results: %v", err)
			return
		}

		if actionResults.Error != nil {
			state.Throw(actionResults.Error.Code, actionResults.Error.Message)
			return
		}

		err = state.Set("return", actionResults.Payload)
		if err != nil {
			state.Throw("direktiv.action.outputUnsaveable", "current state information prevents storing action results at '.return'")
			return
		}

		m := make(map[string]interface{})
		for k, v := range sl.state.Event.Context {
			if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
				jqcmd := v[2:len(v)-2]
				result := state.JQ(jqcmd)
				// TODO: check that result produces a valid type
				m[k] = result
			} else {
				m[k] = v
			}
		}

		err = state.Save([]byte("waitingOnEvent"))
		if err != nil {
			state.Crash(sl, "failed to save state information: %v", err)
			return
		}

		err = state.Yield()
		if err != nil {
			state.Crash(sl, "failed to yield state lock: %v", err)
			return
		}

		err = sl.eventsEngine.UpgradeInterest(sl.InstanceID, sl.namespace, sl.state.Event.Type, m)
		if err != nil {
			state.Crash(sl, "failed to upgrade interest in event: %v", err)
			return
		}

		return

	}

	if step == "waitingOnEvent" {

		eventReceived := new(EventReceived)

		err = json.Unmarshal(wakedata, eventReceived)
		if err != nil {
			state.Crash(sl, "failed to unmarshal event: %v", err)
			return
		}

		err = state.Set(eventReceived.Type, eventReceived.Payload)
		if err != nil {
			state.Throw("direktiv.action.eventUnsaveable", "current state information prevents storing event payload at '.%s'", eventReceived.Type)
			return
		}

		err = state.Transform(sl.state.Transform)
		if err != nil {
			state.Crash(sl, "failed to apply state transform: %v", err)
			return
		}

		err = state.Yield()
		if err != nil {
			state.Crash(sl, "failed to yield state lock: %v", err)
			return
		}

		err = state.Transition(sl.state.Transition)
		if err != nil {
			state.Crash(sl, "failed to yield state lock: %v", err)
			return
		}

		return

	}

	panic(fmt.Errorf("unexpected savedata: %s", step))

}
```




How many ways can Logic be woken up?

- Transitioned into the state. 
- Woken up from a sleep.
- Woken up from an action. 
- Woken up from all actions completing.
- Woken up from any action returning something.
- Woken up from an event. 
- Woken up from all events.
- Woken up from any event.











-Logic needs ability to log something.
-Logic needs ability to throw catchable error. 
-Logic needs ability to run a JQ command.
-Logic needs ability to select the transition.
-Logic needs ability to spin off a thread.
-Logic needs ability to choose whether it must wait for every thread to return or just one. 
-Logic needs ability to optionally override the default error handler.
-Logic needs ability to be cancelled.
-Logic needs ability to be timed out.


Timeouts & Delays (same?), Actions, Register for event, 



Wait: put state onto the database, then kick off other threads. 
	The other threads each need to get the database mutex before returning.
	After getting a lock, the other thread saves its response in its predefined location and marks itself as complete.
	The state logic needs to be able to wake up and check what happened and make decisions.




How to implement a callback?
------------------------------------



How to implement an OR parallel?
------------------------------------


