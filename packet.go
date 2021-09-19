package main

import (
	"context"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

// https://github.com/zh2x/BCI_Protocol

type packet struct {
	ts  int64
	mac bluetooth.MAC
	buf [5]byte
}

func (p *packet) SignalStrength() byte { return p.buf[0] & 7 }
func (p *packet) NoSignal() bool       { return p.buf[0]&(1<<4) > 0 }
func (p *packet) ProbeUnplugged() bool { return p.buf[0]&(1<<5) > 0 }
func (p *packet) PulseBeep() bool      { return p.buf[0]&(1<<6) > 0 }
func (p *packet) Pleth() byte          { return p.buf[1] }
func (p *packet) BarGraph() byte       { return p.buf[2] & 7 }
func (p *packet) NoFinger() bool       { return p.buf[2]&(1<<4) > 0 }
func (p *packet) PulseResearch() bool  { return p.buf[2]&(1<<5) > 0 }
func (p *packet) PulseRate() byte      { return p.buf[3]&127 | (p.buf[2]&(1<<6))<<1 }
func (p *packet) SpO2() byte           { return p.buf[4] }

func (p packet) String() string {
	return fmt.Sprintf("%d,%s,%d,%d,%d,%d,%d", p.ts, p.mac.String(), p.buf[0], p.buf[1], p.buf[2], p.buf[3], p.buf[4])
}

func (p packet) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`{"ts":%d,"ad":"%s","ss":%d,"ns":%v,"un":%v,"bp":%v,"pl":%d,"bg":%d,"nf":%v,"pr":%v,"rate":%d,"spO2",%v}`,
		p.ts, p.mac, p.SignalStrength(), p.NoSignal(), p.ProbeUnplugged(), p.PulseBeep(),
		p.Pleth(), p.BarGraph(), p.NoFinger(), p.PulseResearch(), p.PulseRate(), p.SpO2())), nil
}

func parsePackets(ctx context.Context, mac bluetooth.MAC, uart UART) <-chan packet {
	var (
		st  = 0
		ch  = make(chan packet, 64)
		pkt = packet{mac: mac}
	)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if uart.Buffered() == 0 {
					time.Sleep(time.Millisecond)
					continue
				}
				b, err := uart.ReadByte()
				if err != nil {
					return
				}
				switch st {
				case 0:
					if (1<<7)&b > 0 {
						pkt.buf[st] = b
						st++
					}
				default:
					if (1<<7)&b > 0 {
						st = 0
						continue
					}
					pkt.buf[st] = b
					pkt.ts = time.Now().UnixNano()
					if st == 4 {
						ch <- pkt
						st = 0
					} else {
						st++
						continue
					}
				}
			}
		}
	}()
	return ch
}
