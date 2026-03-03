// Package status 提供了运行时模式、数据头类型和状态常量定义。
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
