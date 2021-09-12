package main

import (
	"context"
	"fmt"
	"sync"

	"tinygo.org/x/bluetooth"
)

func Scan(ctx context.Context, adapter *bluetooth.Adapter, filter ScanResultFilter) *ScanResults {

	const bufferSize = 10

	// internal channel to buffer the scan results
	ch := make(chan bluetooth.ScanResult, bufferSize)
	sr := &ScanResults{
		results: make(chan bluetooth.ScanResult, bufferSize),
		errorch: make(chan error, 2),
	}
	cleanup := func() func() {
		var once sync.Once
		return func() {
			once.Do(func() {
				close(ch)
				if err := adapter.StopScan(); err != nil {
					sr.errorch <- err
				}
				close(sr.errorch)
				close(sr.results)
			})
		}
	}()

	go func() {
		if err := adapter.Scan(func(_ *bluetooth.Adapter, r bluetooth.ScanResult) {
			select {
			case ch <- r: // non-blocking write was successful
			default: // non-blocking write was unsuccessful; ignore this scan result
			}
		}); err != nil {
			sr.errorch <- fmt.Errorf("unable to initiate BLE scan: %w", err)
			cleanup()
		}
	}()

	// filter the scan results from the buffer channel
	go func() {
		defer cleanup()
		cache := make(map[string]bool)
		for {
			select {
			case r := <-ch:
				if r.Address == nil {
					continue
				}
				addr := r.Address.String()
				if _, ok := cache[addr]; ok {
					continue
				}
				if cache[addr] = (filter == nil || filter(r)); cache[addr] {
					sr.results <- r
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return sr
}

type ScanResultFilter func(res bluetooth.ScanResult) bool

func AllScanResults(res bluetooth.ScanResult) bool {
	return true
}

type ScanResults struct {
	res bluetooth.ScanResult
	err error

	results chan bluetooth.ScanResult
	errorch chan error
}

func (sr *ScanResults) Curr() bluetooth.ScanResult {
	return sr.res
}

func (sr *ScanResults) Next() bool {
	select {
	case dev, ok := <-sr.results:
		if !ok {
			return false
		}
		sr.res = dev
		return true
	case err := <-sr.errorch:
		sr.res = bluetooth.ScanResult{}
		sr.err = err
	}
	return false
}

func (sr *ScanResults) Err() error {
	return sr.err
}
