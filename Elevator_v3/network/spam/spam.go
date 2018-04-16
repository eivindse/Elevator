package spam

import (
	"../../types"
)

func AliveSpam(s chan types.ElevatorState, c chan types.ElevatorState) {//Channel chan types.ElevatorState, c chan types.ElevatorState) {
	for {
	state := <- s	
			c <- state
		}
}
func SendOrder(order types.ButtonEvent, oc chan types.ButtonEvent){
	oc <- order
}
func SendOrderToId(order types.ButtonEvent, id int, oc chan types.NewOrder){
	newOrder := types.NewOrder{order, id}
	oc <- newOrder
}

