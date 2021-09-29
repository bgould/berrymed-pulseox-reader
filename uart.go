package main

import (
	"context"
	"fmt"
	"io"

	"github.com/smallnest/ringbuffer"
	"tinygo.org/x/bluetooth"
)

type UART interface {
	io.ByteReader

	Buffered() int
}

func OpenUART(ctx context.Context, device *bluetooth.Device, bufferSize int) (UART, error) {
	svcs, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		return nil, fmt.Errorf("could not discover service: %w", err)
	}
	chars, err := svcs[0].DiscoverCharacteristics([]bluetooth.UUID{charTxUUID, charRxUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to discover RX and TX characteristics: %w", err)
	}
	port := &mchpTransparentUART{dv: device, rb: ringbuffer.New(bufferSize)}
	if err := chars[0].EnableNotifications(port.writeToBuffer); err != nil {
		return nil, fmt.Errorf("could not enable tx notifications: %w", err)
	}
	return port, nil
}

var (
	serviceUUID, _ = bluetooth.ParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
	charTxUUID, _  = bluetooth.ParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")
	charRxUUID, _  = bluetooth.ParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
)

type mchpTransparentUART struct {
	dv *bluetooth.Device
	rb *ringbuffer.RingBuffer
}

//func (port *mchpTransparentUART) Read(buf []byte) (int, error) {
//	return port.rb.Read(buf)
//}

func (port *mchpTransparentUART) ReadByte() (byte, error) {
	return port.rb.ReadByte()
}

func (port *mchpTransparentUART) Buffered() int {
	return port.rb.Length()
}

func (port *mchpTransparentUART) writeToBuffer(value []byte) {
	length := len(value)
	for bytesWritten := 0; bytesWritten < length; {
		n, err := port.rb.Write(value)
		if err != nil {
			return
		}
		bytesWritten += n
		value = value[n:]
	}
}
