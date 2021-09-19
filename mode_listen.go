package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"tinygo.org/x/bluetooth"
)

func listenMode(ctx context.Context, adapter *bluetooth.Adapter) {

	var filter ScanResultFilter
	if addr := address(); addr == "" {
		try(errors.New("address is required"), "invalid argument(s)")
	} else {
		filter = matchAddress(addr)
	}

	for {

		// first perform a bluetooth scan to see if we can find the device
		var foundDevice *bluetooth.ScanResult
		func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			log.Println("scanning...")
			go func() {
				//<-after(2 * time.Second)
				time.Sleep(2 * time.Second)
				cancel()
			}()

			if results := Scan(ctx, adapter, filter); results.Next() {
				result := results.Curr()
				foundDevice = &result
			} else if err := results.Err(); err != nil {
				log.Printf("scanning error: %v", err)
			}
		}()

		// scan didn't turn up the device we're looking for; wait and then try again
		if foundDevice == nil || foundDevice.Address == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// found a matching device; attempt to connect and stream packets from it
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
				case <-after(2 * time.Second):
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
