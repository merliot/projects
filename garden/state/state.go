package state

import "time"

type State struct {
	Msg         string
	Now         time.Time
	StartTime   string
	StartDays   [7]bool
	Gallons     float64
	GallonsGoal uint
	Running     bool
}
