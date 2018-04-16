package master

import (
	"strconv"
	  "../types"
	  "fmt"
)
/*
//This function takes a target floor and a list of available elevators to return the id of the closest one
func getClosestReadyElevatorId(targetFloor int, elevatorList []types.elevatorState) string{
	minDistance := 10000000
	var chosenElevator types.elevatorState

	//Finds the elevator with least distance to the target floor
	for _, elevator := range elevatorList {
		if elevator.ready{
			distance := targetFloor - elevator.Floor
			if distance < 0 {
			distance = -distance
			}
			if distance < minDistance{
				minDistance = distance
				chosenElevator = elevator
			}
		}
	}

	return chosenElevator.id
}
*/
func RemoveFromReadyList(readyList []int, id int) []int{
	fmt.Println("removing elevator from readylist", id)
	var new []int
		for i := range readyList {
			if readyList[i] != id {
				new = append(new, readyList[i])
			}
		}
	return new
}
func RemoveHallOrder(list []types.ButtonEvent, order types.ButtonEvent)  []types.ButtonEvent{
	var new []types.ButtonEvent
		for i := range list {
			if list[i] != order {
				new = append(new, list[i])
			}
		}
	return new
}
func IsHallOrder(list []types.ButtonEvent, order types.ButtonEvent) bool{
  for i := range list{
      if list[i] == order{
          return true
      }
  }
  return false
}
func LowestId(peerList []string) int{
	min := 65535 + 1
	for i := range peerList {
        k, _ := strconv.Atoi(peerList[i])
			if k < min{
					min = k
				}
    }
		return min
}

func ReadyExists(eMap map[int]types.ElevatorState) bool{
	for ele := range eMap{
		if eMap[ele].Ready == true{
			return true
		}
	}
	return false
}

func DelegateOrder(order types.ButtonEvent, eMap map[int]types.ElevatorState, readyList []int) int{ 
	chosenId := -1
	distance := 10000000000
	minDistance := 1000000
	erBtn := types.ButtonEvent{0,2}
	for i := range readyList{
		if eMap[readyList[i]].DelegatedOrder == erBtn{
			distance = order.Floor - eMap[readyList[i]].Floor
			if (distance < 0){distance = distance * -1}

			if distance < minDistance{
				minDistance = distance
				chosenId = readyList[i]
			}
		}
	}
	//ferdgigjør
	if chosenId != -1{
		//fmt.Println("Sending order to ", chosenId2)
		return chosenId
	}
	return 0
}
func IsDelegated(order types.ButtonEvent, delegated []types.ButtonEvent) bool{
	for e := range delegated{
		if delegated[e] == order{
			return true
		}
	}
	return false
}
func NonDelegatedExists(orders []types.ButtonEvent, delegated []types.ButtonEvent) bool{
	for o := range orders{
		for e := range delegated{
			if orders[o] != delegated[e]{
				return true
			}
		}
	}
	return false
}
func GetFreeOrder(orders []types.ButtonEvent, occupied []types.ButtonEvent) types.ButtonEvent{
	if len(occupied) != 0{
		for o := range orders{
			for e := range occupied{
				if orders[o] != occupied[e]{
					return orders[o]
				}
			}
		}
		errorO := types.ButtonEvent{0,2}
		return errorO
	}else{
		return orders[0]
	}
}
/*
func removeHallOrder(a types.ButtonEvent){
	var new []types.ButtonEvent
	elevio.SetButtonLamp(a.Button, a.Floor, false)
	  for i := range hallOrders {
		  if hallOrders[i] != a {
			  new = append(new, hallOrders[i])
		  }
	  }
	hallOrders = new
	openDoor()
	getNewTarget(a.Floor)
	removingHall = true
	lastOrderCompleted = a
	state.DoneOrders = a
}
func addHallOrder(a types.ButtonEvent){
	var dir types.MotorDirection
	if a.Button == 0{
	  dir = types.MD_Up
	}else{
	  dir = types.MD_Down
	}
	if !isHallOrder(a.Floor, dir){
	  hallOrders = append(hallOrders, a)
	}
	if target == -1 || target == currentFloor{
	  getNewTarget(currentFloor)
	}
}*/