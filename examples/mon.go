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

	err = cib.Subscribe(func(event pacemaker.CibEvent, cib string) {
		if event == pacemaker.UpdateEvent {
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
		fmt.Printf("waiting...\n")
		time.Sleep(5*time.Second)
	}
}
