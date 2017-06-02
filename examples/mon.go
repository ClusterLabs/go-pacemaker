package main

import (
	"fmt"
	"log"
	"time"
	"github.com/krig/go-pacemaker"
)

func listenToCib(cib *pacemaker.Cib, restarter chan int) {
	err := cib.Subscribe(func(event pacemaker.CibEvent, cib string) {
		if event == pacemaker.UpdateEvent {
			fmt.Printf("\n")
			fmt.Printf("event: %s\n", event)
			fmt.Printf("cib: %s\n", cib)
		} else {
			log.Printf("lost connection: %s\n", event)
			restarter <- 1
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to CIB: %s", err)
	}
}

func connectToCib() (*pacemaker.Cib, error) {
	cib, err := pacemaker.OpenCib()
	if err != nil {
		log.Print("Failed to open CIB")
		return nil, err
	}

	xmldata, err := cib.Query()
	if err != nil {
		log.Print("Failed to query CIB")
		return nil, err
	}
	fmt.Printf("output: %s\n", xmldata)
	return cib, nil
}

func main() {	
	restarter := make(chan int)

	cib, err := connectToCib()
	if err != nil {
		log.Printf("Failed in connectToCib: %s", err)
	} else {
		listenToCib(cib, restarter)
	}
	go pacemaker.Mainloop()
	state := 0
	for {
		if state == 0 {
			state = <-restarter
		} else if state == 1 {
			if cib != nil {
				cib.Close()
				cib = nil
			}
			cib, err = connectToCib()
			if err != nil {
				log.Printf("Failed in connectToCib: %s", err)
				time.Sleep(5 * time.Second)
			} else {
				listenToCib(cib, restarter)
				state = 0
			}
		}
	}
}
