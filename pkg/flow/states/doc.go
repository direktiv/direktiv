// Package states implements the logic for Direktiv workflow states.
//
// State logic is implemented in Run functions.
//
// The following state properties should be ignored in this package because they're handled by Direktiv:
// * catch
// * log
// * metadata
//
// The state logic dictates how to transform and transition with the returned Transition struct. If the
// Transition struct is not nil Direktiv will execute the transform and transition logic. As long as the
// transition string is not empty it must refer to a state ID. If it's empty, that concludes the workflow.
// If the Run function returns a nil Transition struct and no errors, this tells Direktiv to yield. This
// is how long-running states can be implemented: they yield, putting the instance to sleep while it waits
// for something to wake it up. Such states usually should call Engine APIs to register for ways that they
// can be woken up.
//
// When Direktiv first executes a state it calls the Run function with nil wakedata and the instance will
// have nothing stored in memory. If the state may need to be scheduled in repeatedly, the Run function
// will need to determine where it stands in that process in any given call by checking what's in instance
// memory and the wakedata. The wakedata is defined in various ways by the Engine APIs, but the instance
// memory is entirely under the control of the Run logic.
//
// Run functions should be designed in a way such that each time they are called they will return in a
// timely manner. The entire time the Run function is running Direktiv holds a cluster-wide mutex on the
// instance, making it impossible to cancel or timeout. Ensure that this is always only temporary. Even
// though the mutex is on the specific instance, each node in the cluster can only hold a limited number
// of cluster-wide mutexes at any given time. Direktiv uses these cluster-wide locks for many purpose,
// therefore, failure to return in a timely manner can result in seemingly unrelated problems in the
// cluster.
//
// The mutex held for an instance guarantees that exactly one node in the cluster can be running logic
// for the state at a time. As an example: it is safe to register an event listener in a Run function
// without worrying about race conditions because even if the events are received immediately the
// followup call to Run will wait until after the current call has returned and been cleaned up.
package states
