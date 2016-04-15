package queue

import (
	def "definitions"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// RunBackup loads backup on startUp, and saves queue whenever
// there is anything on the takeBackup channel
func RunBackup(outgoingMsg chan<- def.Message) {

	const filename = "elevator_backup.dat"
	var backup QueueType
	backup.loadFromDisk(filename)
	printQueue()
	// Read last time backup was modified
	fileStat, err := os.Stat(filename)
	if err != nil {
		log.Println(def.ColR, err, def.ColN)
	}

	// Resend all hall requests found in backup, and add cab requests to queue:
	for floor := 0; floor < def.NumFloors; floor++ {
		for btn := 0; btn < def.NumButtons; btn++ {
			if backup.hasRequest(floor, btn) {
				if btn == def.BtnCab {
					AddRequest(floor, btn, def.LocalIP)
				} else if time.Now().After(fileStat.ModTime().Add(def.RequestTimeoutDuration)) {
					outgoingMsg <- def.Message{Category: def.NewRequest, Floor: floor, Button: btn}
				}
			}
		}
	}
	go func() {
		for {
			<-takeBackup
			log.Println(def.ColG, "Take Backup", def.ColN)
			if err := queue.saveToDisk(filename); err != nil {
				log.Println(def.ColR, err, def.ColN)
			}
		}
	}()
}

// saveToDisk saves a QueueType to disk.
func (q *QueueType) saveToDisk(filename string) error {

	data, err := json.Marshal(&q)
	//log.Println(string(data))
	if err != nil {
		log.Println(def.ColR, "json.Marshal() error: Failed to backup.", def.ColN)
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Println(def.ColR, "ioutil.WriteFile() error: Failed to backup.", def.ColN)
		return err
	}
	return nil
}

// loadFromDisk checks if a file of the given name is available on disk, and
// saves its contents to a QueueType
func (q *QueueType) loadFromDisk(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		log.Println(def.ColG, "Backup file found, processing...", def.ColN)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(def.ColR, "loadFromDisk() error: Failed to read file.", def.ColN)
		}
		if err := json.Unmarshal(data, q); err != nil {
			log.Println(def.ColR, "loadFromDisk() error: Failed to Unmarshal.", def.ColN)
		}
	}
	return nil
}

func printQueue() {
	fmt.Println(def.ColB, "\n*****************************")
	fmt.Println("*       Up     Down    Cab   ")
	for f := def.NumFloors - 1; f >= 0; f-- {
		s := "* " + strconv.Itoa(f+1) + "  "
		for b := 0; b < def.NumButtons; b++ {
			if queue.hasRequest(f, b) && b != def.BtnCab {
				s += "( " + queue.Matrix[f][b].Addr[12:15] + " ) "
			} else if queue.hasRequest(f, b) {
				s += "(  x  ) "
			} else {
				s += "(     ) "
			}
		}
		fmt.Println(s)
	}
	fmt.Println("*****************************\n", def.ColN)
}
