// Package job provides a robust, testable job queue with a bounded work channel
// and a fixed worker pool. It is designed to execute CPU-bound tasks like
// running Sobek/goja scripts safely with context deadlines, panic containment,
// and backpressure.
//
// Integration notes (for your engine):
//   - Create a Manager with NewManager and start it with Start().
//   - Enqueue jobs from your ExecScript endpoint using Enqueue().
//   - Provide a Runner implementation that executes your JS (one VM per job).
//   - Use context deadlines/timeouts per job for interruption/cancellation.
package job
