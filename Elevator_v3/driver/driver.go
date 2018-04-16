package driver

import "./elevio"
import "fmt"
import "../types"
import "../config"
import "../dataStore"
import "time"
import "strconv"

var state types.ElevatorState
var target int
var cabOrders []int
var hallOrders []types.ButtonEvent
var direction types.MotorDirection
var toggle = true
var obstruction bool
var door bool
//Cab
func isCabOrder(floor int) bool{ //is there a cab order in this floor
  for i := range cabOrders{
      if cabOrders[i] == floor{
          return true
      }
  }
  return false
}
func addCabOrder(floor int){//Add a cab order to package list
for i := range cabOrders{
  if cabOrders[i] == floor{
    return
  }
}
cabOrders = append(cabOrders, floor)
}
func removeCabOrder(floor int){//Remove a cab order form package list
var newCabOrderList []int
for order := range cabOrders{
  if cabOrders[order] != floor{
    newCabOrderList = append(newCabOrderList, cabOrders[order])
  }else{
    elevio.SetButtonLamp(2, floor, false)
  }
}
cabOrders = newCabOrderList
}
func updateCabLights(){//Updates lights for simulator
for order := range cabOrders{
  elevio.SetButtonLamp(2, cabOrders[order],  true)
}

}
func findClosestCab(floor int) int{
  var targetFloor int
  minimum := 10000000
  for i := range cabOrders{
      k := floor - cabOrders[i]
      if k < 0 { k = -k }

      if k < minimum{
          minimum = k
          targetFloor = cabOrders[i]
      }
  }
  return targetFloor
}

//Hall
func isHallOrder(floor int, dir types.MotorDirection) bool{ //is there a hall order in this floor
  var btn types.ButtonType
  if dir == types.MD_Down{
    btn = 1
  }else{
    btn = 0
  }

  for i := range hallOrders{
      if hallOrders[i].Floor == floor && hallOrders[i].Button == btn{
          return true
      }
  }
  return false
}

func openDoor(){
  target = -1
  elevio.SetDoorOpenLamp(true)
  door = true
  time.Sleep(3* time.Second)
  elevio.SetDoorOpenLamp(false)
  door = false
  getNewTarget(state.Floor)
  engineUpdate(state.Floor)
}

func floorUpdate(floor int, completedOrderChannel chan types.ButtonEvent){
  newTarget := target
  pickUp := false
  target = -1
  //engineUpdate(floor)
  
  var bt types.ButtonType
  if direction == types.MD_Down{bt = 1}else{bt = 0}
  
  order := types.ButtonEvent{floor, bt}

  if isCabOrder(floor){
    pickUp = true
  }else if isHallOrder(floor, direction){
    pickUp = true
  }
  if floor == config.NumFloors-1 && isHallOrder(floor, types.MD_Down){
    order = types.ButtonEvent{floor, 1}
    pickUp = true
  }else if floor == 0 && isHallOrder(floor, types.MD_Up){
    order = types.ButtonEvent{floor, 0}
    pickUp = true
  }
  o1 := types.ButtonEvent{floor, 0}
  o2 := types.ButtonEvent{floor, 1}
  if state.DelegatedOrder == o1{
    order = o1
    state.DelegatedOrder = types.ButtonEvent{0,2}
    pickUp = true
  }else if state.DelegatedOrder == o2{
    order = o2
    state.DelegatedOrder = types.ButtonEvent{0,2}
    pickUp = true
  }
  if pickUp {
    go openDoor()
    go func(){completedOrderChannel <- order}()
    removeCabOrder(floor)
  }
  if !door {
    if floor == newTarget{getNewTarget(floor)}else if target == -1{getNewTarget(floor)}
    engineUpdate(floor)
  }
}
func getNewTarget(floor int){
  if obstruction || door{
    target = -1
  }else if len(cabOrders) != 0{
      target = findClosestCab(floor)
  }else if state.DelegatedOrder.Button != 2{
      target = state.DelegatedOrder.Floor
  }else{
      target = -1
  }
}
func updateHallLights(){//Updates lights for simulator
  elevio.TurnOffLights(config.NumFloors)
  for _, order := range hallOrders{
    elevio.SetButtonLamp(order.Button, order.Floor, true)
  }
}
func updateLights(){
  updateCabLights()
  updateHallLights()
}
func engineUpdate(floor int){
  if target == -1{
    direction = types.MD_Stop
  }else if floor < target{
    direction = types.MD_Up
  }else if floor > target{
    direction = types.MD_Down
  }else{
    direction = types.MD_Stop
  }
  updateLights()
  elevio.SetMotorDirection(direction)
}
//main goroutine
func DriverMain(buttonChannel chan types.ButtonEvent, stateChannel chan types.ElevatorState, initState types.ElevatorState, completedOrderChannel chan types.ButtonEvent, hallOrderChannel chan []types.ButtonEvent, doOrderChannel chan types.ButtonEvent, nowReadyChannel chan int){
  state = initState
  obstruction = false
  target = -1
  direction = types.MD_Down
  door = false

  filename := fmt.Sprintf("dataStore/%s", strconv.Itoa(state.Id))
  if dataStore.DoesFileExist(filename){
    cabOrders = dataStore.LoadList(state.Id)
  }else{
    dataStore.CreateFile(filename)
  }
  
  //Laste inn cab orders fra fil

  np := fmt.Sprintf("localhost:%d", state.Id)
  elevio.Init(np, config.NumFloors)
  elevio.SetMotorDirection(types.MD_Down)
  
  drv_buttons := make(chan types.ButtonEvent)
  drv_floors  := make(chan int)
  drv_obstr   := make(chan bool)
  drv_stop    := make(chan bool)

  go elevio.PollButtons(drv_buttons)
  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollObstructionSwitch(drv_obstr)
  go elevio.PollStopButton(drv_stop)
  updateLights()
  for {
    if (target == -1){state.Ready = true}else{state.Ready = false}
    select {
      case button := <- drv_buttons: //button pressed
        if button.Button != 2{
          go func(){buttonChannel <- button}()
        }else{
          addCabOrder(button.Floor)
        }
        if direction == types.MD_Stop{
          floorUpdate(state.Floor, completedOrderChannel)
        }
        fmt.Println(cabOrders)
        updateLights()
        engineUpdate(state.Floor)
      case order := <- doOrderChannel:
        fmt.Println("Fikk ordre ", order, " fra nettverk!")
        if state.DelegatedOrder.Button == 2{
          state.DelegatedOrder = order
        }else if order.Button == 2{
          state.DelegatedOrder = order
        }else{
          fmt.Println("Prøver å delegere order selv om heis har et allerede...")
        }
        getNewTarget(state.Floor)
        engineUpdate(state.Floor)
        toggle = true
      case list := <- hallOrderChannel:
        hallOrders = list
        
      case floor := <- drv_floors: //ny etasje
        state.Floor = floor
        getNewTarget(state.Floor)
        floorUpdate(floor, completedOrderChannel)
        engineUpdate(floor)
      case a := <- drv_obstr: //obstruction
          fmt.Printf("OBSTR: %+v\n", a)
          obstruction = a
          getNewTarget(state.Floor)
          engineUpdate(state.Floor)
      case a := <- drv_stop:
          fmt.Printf("STOP: %+v\n", a)
          for f := 0; f < config.NumFloors; f++ {
              for b := types.ButtonType(0); b < 3; b++ {
                  elevio.SetButtonLamp(b, f, false)
              }
          }
      case <- time.After(500 * time.Millisecond):
        updateLights()
        go func(){stateChannel <- state}()
        dataStore.SaveList(cabOrders, state.Id)  
      }
      //Oppdater state til main
  }
}
