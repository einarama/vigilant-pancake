package main

import (
	def "definitions"
	hw "hardware"
)


func EventHandler(){
	//Check for all events in loop
	//Make convinient variables
	//Fix lights
	
	//Convenient channels
	var BtnChan = make(chan def.BtnPress, 10)


	//Convenient variables/structs


	//Threads
	go eventBtnPressed(BtnChan)


	for{
		select{
			case
		}
	}
}

func eventBtnPressed(ch chan){
	//Check for a button beeing pressed
	lastBtnPressed := def.BtnPress{
		Button: -1,
		Floor: -1,
	} 
	btnPressed := def.BtnPress{
		Button: -2,
		Floor: -2,
	}
	for{
		for floor := 0; floor < def.NumFloors; floor++ {
			for btn := 0; btn < def.NumButtons; btn++ {
				if hw.ReadBtn(floor, btn){
					btnPressed.Button = btn
					btnPressed.Floor = floor
					//Check if there is an order assigned to button
					if lastBtnPressed != btnPressed && !qMatrix.Active{
						ch <- btnPressed
					}
				}
			}
		}
		lastBtnPressed = btnPressed
	}
}