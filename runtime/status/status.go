// Package status provides runtime mode, data header type, and status constant definitions.
package status

// Runtime mode constants.
const (
	UNSET int = iota - 1
	OFFLINE
	SERVER
	CLIENT
)

// Data header type constants.
const (
	REQTASK = iota + 1 // task request header
	TASK               // task response stream header
	LOG                // log output header
)

// Runtime status constants.
const (
	STOPPED = iota - 1
	STOP
	RUN
	PAUSE
)
