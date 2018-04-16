package types

type AliveMsg struct {
	Name string
}


type ElevatorState struct {
	Ip string
	Id int
	Floor int
	Ready bool
	DelegatedOrder ButtonEvent
	//Direction??????
}

type NewOrder struct {
	Button ButtonEvent
	Id int
}

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)


type ButtonEvent struct {
	Floor  int
	Button ButtonType
}
