package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/smallnest/ringbuffer"
	"tinygo.org/x/bluetooth"
)

var (
	serviceUUID = mustParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
	charTxUUID  = mustParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")
	charRxUUID  = mustParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
)

func mustParseUUID(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

var (
	ErrPortClosed      = errors.New("port is closed")
	ErrServiceNotFound = errors.New("service not found")
)

type Port struct {
	device  *bluetooth.Device
	service bluetooth.DeviceService
	charTx  bluetooth.DeviceCharacteristic
	charRx  bluetooth.DeviceCharacteristic

	rbuf *ringbuffer.RingBuffer

	mutex  sync.RWMutex
	closed bool
}

func (port *Port) Read(buf []byte) (int, error) {
	port.mutex.RLock()
	if port.closed {
		return 0, ErrPortClosed
	}
	defer port.mutex.RUnlock()
	return port.rbuf.Read(buf)
}

func (port *Port) ReadByte() (byte, error) {
	port.mutex.RLock()
	if port.closed {
		return 0, ErrPortClosed
	}
	defer port.mutex.RUnlock()
	return port.rbuf.ReadByte()
}

func (port *Port) Write(buf []byte) (int, error) {
	return 0, errors.New("not implemented yet")
}

func (port *Port) WriteByte(b byte) error {
	return errors.New("not implemented yet")
}

func HasService(adapter *bluetooth.Adapter, addresser bluetooth.Addresser) bool {
	device, err := adapter.Connect(addresser, bluetooth.ConnectionParams{})
	if err != nil {
		return false
	}
	defer device.Disconnect()
	svcs, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil || len(svcs) == 0 {
		return false
	}
	return true
}

func OpenPort(adapter *bluetooth.Adapter, addresser bluetooth.Addresser, timeout time.Duration) (*Port, error) {
	device, err := adapter.Connect(addresser, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("could not connect: %w", err)
	}
	svcs, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		return nil, fmt.Errorf("could not discover service: %w", err)
	}
	if len(svcs) == 0 {
		return nil, ErrServiceNotFound
	}
	service := svcs[0]
	chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{charRxUUID, charTxUUID})
	if err != nil {
		return nil, fmt.Errorf("Failed to discover RX and TX characteristics: %w", err)
	}
	port := &Port{
		device:  device,
		service: service,
		charRx:  chars[0],
		charTx:  chars[1],
		rbuf:    ringbuffer.New(1024),
	}
	writeToBuffer := func(value []byte) {
		for bytesWritten := 0; bytesWritten < len(value); {
			n, err := port.rbuf.Write(value)
			if err != nil {
				return
			}
			bytesWritten += n
			value = value[n:]
		}
	}
	time.Sleep(5 * time.Second)
	// Enable notifications to receive incoming data.
	if err := port.charTx.EnableNotifications(writeToBuffer); err != nil {
		port.Close()
		return nil, fmt.Errorf("could not enable tx notifications: %w", err)
	}
	return port, nil
}

func (port *Port) Close() error {
	port.mutex.Lock()
	defer port.mutex.Unlock()
	if !port.closed {
		err := port.device.Disconnect()
		port.closed = true
		return err
	}
	return nil
}
