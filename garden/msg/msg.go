package msg

type Update struct {
	Msg     string
	Gallons float64
	Running bool
}

type Day struct {
	Msg   string
	Day   uint
	State bool
}

type StartTime struct {
	Msg   string
	Time  string
}
