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
		var foundDevice *bluetooth.ScanResult

		func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				time.Sleep(2 * time.Second)
				cancel()
			}()
			log.Println("scanning...")

			if results := Scan(ctx, adapter, matchAddress(address)); results.Next() {
				result := results.Curr()
				foundDevice = &result
			} else if err := results.Err(); err != nil {
				log.Printf("scanning error: %v", err)
			}
		}()

		if foundDevice == nil || foundDevice.Address == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		func() {

			defer time.Sleep(time.Second)

			addrstr := foundDevice.Address.String()
			addr, _ := bluetooth.ParseMAC(addrstr)

			device, err := adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{})
			if err != nil {
				log.Printf("could not connect: %v", err)
				return
			}
			defer device.Disconnect()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			port, err := OpenUART(ctx, device, 1024)
			if err != nil {
				log.Printf("failed to open UART %s: %v", foundDevice.Address, err)
				return
			}

			for packets := parsePackets(ctx, addr, port); ; {
				select {
				case packet := <-packets:
					//pkt, _ := packet.MarshalJSON()
					pkt := packet.String()
					os.Stdout.Write([]byte(string(pkt) + "\n"))
					continue
				case <-time.After(2 * time.Second):
					cancel()
					return
				case <-ctx.Done():
					return
				}
			}

		}()

	}

}

func matchAddress(addr string) func(bluetooth.ScanResult) bool {
	return func(res bluetooth.ScanResult) bool {
		if res.Address == nil {
			return false
		}
		return res.Address.String() == addr
	}
}
