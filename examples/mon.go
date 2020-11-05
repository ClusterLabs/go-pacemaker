package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ClusterLabs/go-pacemaker"
)

var verbose = flag.Bool("verbose", false, "print whole cib on each update")
var file = flag.String("file", "", "file to load as CIB")
var remoteSrv = flag.String("remote", "", "remote server to connect to (ip)")
var port = flag.Int("port", 3121, "remote port to connect to (3121)")
var user = flag.String("user", "hacluster", "remote user to connect as")
var password = flag.String("password", "", "remote password to connect with")
var encrypted = flag.Bool("encrypted", false, "set if remote connection is encrypted")

func listenToCib(cib *pacemaker.Cib, restarter chan int) {
	_, err := cib.Subscribe(func(event pacemaker.CibEvent, doc *pacemaker.CibDocument) {
		if event == pacemaker.UpdateEvent {
			fmt.Printf("\n")
			fmt.Printf("event: %s\n", event)
			if *verbose {
				fmt.Printf("cib: %s\n", doc.ToString())
			}
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
	var cib *pacemaker.Cib
	var err error
	// connect to CIB in various ways
	// 1 using a file
	if *file != "" {
		cib, err = pacemaker.OpenCib(pacemaker.FromFile(*file))
		// 2 using a remote server
	} else if *remoteSrv != "" {
		cib, err = pacemaker.OpenCib(pacemaker.FromRemote(*remoteSrv, *user, *password, *port, *encrypted))
		// 3 assuming cib is local to the binary
	} else {
		cib, err = pacemaker.OpenCib()
	}
	if err != nil {
		log.Print("Failed to open CIB")
		return nil, err
	}

	doc, err := cib.Query()
	if err != nil {
		log.Print("Failed to query CIB")
		return nil, err
	}
	defer doc.Close()
	if *verbose {
		fmt.Printf("CIB: %s\n", doc.ToString())
	}
	return cib, nil
}

func main() {
	flag.Parse()

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
