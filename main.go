package main

import (
	"context"
	"log"
	"os"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var address string

func try(err error, msg string) {
	if err != nil {
		for {
			log.Println(msg, ": ", err.Error())
			time.Sleep(2 * time.Second)
		}
	}
}

func main() {

	wait()
	log.Println("starting")

	try(adapter.Enable(), "unable to initialize bluetooth stack")

	log.Println("initialized bluetooth stack")
	if address == "" {
		for {
			println("address is required")
			time.Sleep(1 * time.Second)
		}
	}

	for {

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()
		log.Println("scanning...")
		devs, errs := Scan(ctx, adapter, func(res bluetooth.ScanResult) bool {
			if res.Address == nil {
				return false
			}
			return res.Address.String() == address
		})

		var foundDevice *bluetooth.ScanResult
		select {
		case dev := <-devs:
			foundDevice = &dev
			cancel()
		case err := <-errs:
			log.Printf("error scanning: %v", err)
		}

		if foundDevice == nil || foundDevice.Address == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		port, err := OpenPort(adapter, foundDevice.Address, 0)
		if err != nil {
			log.Printf("failed to open port %s: %v", foundDevice.Address, err)
			time.Sleep(1 * time.Second)
			continue
		}

		for packet := range parse(port.rbuf, 3*time.Second) {
			os.Stdout.Write([]byte(packet.String() + "\n"))
		}

	}

}
