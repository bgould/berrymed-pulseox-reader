package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"tinygo.org/x/bluetooth"
)

type ScanResultFilter func(res bluetooth.ScanResult) bool

func AllScanResults(res bluetooth.ScanResult) bool {
	return true
}

func Scan(
	ctx context.Context,
	adapter *bluetooth.Adapter,
	filter ScanResultFilter) (<-chan bluetooth.ScanResult, <-chan error) {

	const bufferSize = 10

	var (
		results = make(chan bluetooth.ScanResult, bufferSize)
		errorch = make(chan error, 1)
	)

	var (
		// internal channel to buffer the scan results
		scanchan = make(chan bluetooth.ScanResult, bufferSize)
		// since cleanup function closes channels, make sure it is only called once
		once    sync.Once
		cleanup = func() {
			once.Do(func() {
				close(scanchan)
				if err := adapter.StopScan(); err != nil {
					errorch <- err
				}
				close(errorch)
			})
		}
	)
	go func() {
		if err := adapter.Scan(func(_ *bluetooth.Adapter, r bluetooth.ScanResult) {
			select {
			case scanchan <- r: // non-blocking write was successful
			default: // non-blocking write was unsuccessful; ignore this scan result
			}
			runtime.Gosched()
		}); err != nil {
			errorch <- fmt.Errorf("unable to initiate BLE scan: %w", err)
			cleanup()
		}
	}()

	// filter the scan results from the buffer channel
	go func() {
		cache := make(map[string]bool)
		defer func() {
			close(results)
			cleanup()
		}()
		for {
			select {
			case r := <-scanchan:
				if r.Address == nil {
					continue
				}
				addr := r.Address.String()
				if _, ok := cache[addr]; ok {
					continue
				}
				if cache[addr] = (filter == nil || filter(r)); cache[addr] {
					results <- r
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, errorch
}
