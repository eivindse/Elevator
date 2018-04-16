package main

import (
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./network/spam"
	"./types"
	"./driver"
	"./master"
	"strconv"
	"flag"
	"fmt"
	"os"
	"time"
)
var state types.ElevatorState
var isMaster bool //er denne heisen master?
var masterHallOrders []types.ButtonEvent //master sin liste av hallorders
var elevatorMap = make(map[int]types.ElevatorState) //map med heisene
var readyList []int
var peerList []string
var updateCount int
var delegatedList []types.ButtonEvent
var timerMap = make(map[types.ButtonEvent]int)

func main() {
	//initialisering
	state = types.ElevatorState{"",	0, 0, false, types.ButtonEvent{0,2}}
	updateCount = 0

	flag.IntVar(&state.Id, "id", 0, "id of this peer")
	flag.Parse()

	if state.Ip == ""{ //hvis IP ikke er gitt, gi den local ip
		ip_tmp, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			state.Ip = "DISCONNECTED"
		}
		state.Ip = ip_tmp
	}
	if state.Id == 0 { //Hvis id ikke er gitt, gi den prosessid
		state.Id, _ = strconv.Atoi(fmt.Sprintf("%d", os.Getpid()))
	}


	//Updatekanaler
	ticker := time.NewTicker(1 * time.Second)

	//Kanaler mellom main og driver
	buttonChannel := make(chan types.ButtonEvent) //Når en knapp trykkes
	stateChannel := make(chan types.ElevatorState) //Heisens state
	completedOrderChannel:= make(chan types.ButtonEvent) //Når et hallorder er ferdiggjort
	hallOrderChannel := make(chan []types.ButtonEvent) //Hele listen av hallorders for oppdatering av lys og når heisen kjører forbi en etasje
	doOrderChannel := make(chan types.ButtonEvent) //Når det kommer beskjed fra master om å gjøre en ordre
	nowReadyChannel := make(chan int) //Når en heis blir klar sender den sin ID
	spamChannel := make(chan types.ElevatorState) //Spamstate

	//Nettverkskanaler
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	aliveT := make(chan types.ElevatorState) //Sender "Jeg lever" med info
	aliveR := make(chan types.ElevatorState)
	completedOrderT := make(chan types.ButtonEvent) //Sender ut når den har utført en ordre
	completedOrderR := make(chan types.ButtonEvent)
	newOrderT := make(chan types.ButtonEvent) //Sender ut når en heis mottar et knappetrykk
	newOrderR := make(chan types.ButtonEvent)
	nowReadyT := make(chan int) //Sender ut sin ID når den blir klar
	nowReadyR := make(chan int) //Master mottar ID til en heis som blir klar
	orderListR := make(chan []types.ButtonEvent) //Master sender ut oppdater ordreliste
	orderListT := make(chan []types.ButtonEvent)
	doOrderT := make(chan types.NewOrder) //Sender ut et ordre som skal utføres
	doOrderR := make(chan types.NewOrder)

	//Oppretter Transmitter/Receiver-par
	go peers.Transmitter(15647, strconv.Itoa(state.Id), peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16560, aliveT)
	go bcast.Receiver(16560, aliveR)
	go bcast.Transmitter(16561, completedOrderT)
	go bcast.Receiver(16561, completedOrderR)
	go bcast.Transmitter(16562, newOrderT)
	go bcast.Receiver(16562, newOrderR)
	go bcast.Transmitter(16563, nowReadyT)
	go bcast.Receiver(16563, nowReadyR)
	go bcast.Transmitter(16564, orderListT)
	go bcast.Receiver(16564, orderListR)
	go bcast.Transmitter(16565, doOrderT)
	go bcast.Receiver(16565, doOrderR)

	//goroutines
	go driver.DriverMain(buttonChannel, stateChannel, state, completedOrderChannel, hallOrderChannel, doOrderChannel, nowReadyChannel)
	go spam.AliveSpam(spamChannel, aliveT)

	for {
		//fmt.Println("master hall orders: ", masterHallOrders)
		select {
		//nettverk
		case p := <-peerUpdateCh:
			peerList = p.Peers
			isMaster = (master.LowestId(peerList) == state.Id)
			if isMaster{
				fmt.Println("Jeg er master")
			}
			for e := range elevatorMap{
				delete(elevatorMap, e)
			}
		case a := <- aliveR:
			elevatorMap[a.Id] = a
		case o := <- completedOrderR: //får melding fra heis om at ordre er gjort
		fmt.Println("Completed: ", o)
			if isMaster{
				delete(timerMap, o)
				delegatedList = master.RemoveHallOrder(delegatedList, o)
				fmt.Println("For sletting master", masterHallOrders)
				masterHallOrders = master.RemoveHallOrder(masterHallOrders, o)
				fmt.Println("Etter sletting master", masterHallOrders)
				for e := range elevatorMap{
					if elevatorMap[e].DelegatedOrder == o{
						noBtn := types.ButtonEvent{0,2}
						noOrder := types.NewOrder{noBtn, elevatorMap[e].Id}
						go func(){doOrderT <- noOrder}()
					}
				}
			}
		case n := <- newOrderR: //nytt ordre i listen
			if isMaster && !master.IsHallOrder(masterHallOrders, n){
				masterHallOrders = append(masterHallOrders, n)
				go func(){orderListT <- masterHallOrders}()
			}
		case o := <- doOrderR: //Får et nytt order fra master
			if o.Id == state.Id{
				go func(){doOrderChannel <- o.Button}()
			}
		case list := <- orderListR: //mottar liste av hallorders
			if len(masterHallOrders) == 0{
				masterHallOrders = list
			}else if !isMaster{
				masterHallOrders = list
			}

		//driver
		case button := <- buttonChannel: //Sender ut buttonpress til nettverk
			go func(){newOrderT <- button}()
		case s := <- stateChannel: //Når driver oppdaterer state
			go func(){spamChannel <- state}()
			state = s
		case order := <- completedOrderChannel: //Når driver har utført et ordre
			go func(){completedOrderT <- order}()
		case id := <- nowReadyChannel:
			go func(){nowReadyT <- id}()
		

		//updating
		case <- ticker.C: //oppdaterer hvert sekund
		fmt.Println("Masterliste: ", masterHallOrders)
			for ot := range timerMap{
				timerMap[ot]++
				if timerMap[ot] >= 10{
					fmt.Println("ORDER LOST: ", ot)
					delete(timerMap, ot)
					delegatedList = master.RemoveHallOrder(delegatedList, ot)
					if !master.IsHallOrder(masterHallOrders, ot){
						masterHallOrders = append(masterHallOrders, ot)
					}
				}
			}
		if isMaster && len(readyList) != 0 && len(masterHallOrders) != 0 {//} && master.NonDelegatedExists(masterHallOrders, delegatedList){ //delegerer ordre hvis ordre finnes
			button := master.GetFreeOrder(masterHallOrders, delegatedList)
			cid := master.DelegateOrder(button, elevatorMap, readyList)
			order := types.NewOrder{button, cid}
			if  elevatorMap[cid].DelegatedOrder.Button == 2 && button.Button != 2{
				fmt.Println("Delegerer ", order, " til ", cid)
				delegatedList = append(delegatedList, button)
				timerMap[button] = 0
				go func(){doOrderT <- order}()
			}
		}
		readyList = nil
		for i := range elevatorMap{
			if elevatorMap[i].Ready && elevatorMap[i].DelegatedOrder.Button == 2{
				readyList = append(readyList, elevatorMap[i].Id)
			}
		}
		go func(){hallOrderChannel <- masterHallOrders}()
		go func(){orderListT <- masterHallOrders}()
		}
	}
}