package main

import (
	"fmt"
	"log"
	"time"
	"github.com/krig/go-pacemaker"
)

func main() {
	cib, err := pacemaker.OpenCib()
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	func() {
		xmldata, err := cib.Query()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("output: %s\n", xmldata)
	}()

	err = cib.Subscribe(func(event pacemaker.CibEvent, cib string) {
		if event == pacemaker.UpdateEvent {
			fmt.Printf("\n")
			fmt.Printf("event: %s\n", event)
			fmt.Printf("cib: %s\n", cib)
		} else {
			// lost connection!
			// need to schedule a reconnect somehow...
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	
	go func() {
		pacemaker.Mainloop()
	}()
	for {
		fmt.Printf(".")
		time.Sleep(5*time.Second)
	}
}
