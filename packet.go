package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/smallnest/ringbuffer"
)

type packet struct {
	ts  time.Time
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
	return fmt.Sprintf(
		`{"ts":%d,"ss":%d,"ns":%v,"un":%v,"bp":%v,"pl":%d,"bg":%d,"nf":%v,"pr":%v,"rate":%d,"spO2",%v}`,
		p.ts.UnixNano(), p.SignalStrength(), p.NoSignal(), p.ProbeUnplugged(), p.PulseBeep(),
		p.Pleth(), p.BarGraph(), p.NoFinger(), p.PulseResearch(), p.PulseRate(), p.SpO2())
}

func parse(rb *ringbuffer.RingBuffer, timeout time.Duration) <-chan packet {
	st := 0
	ch := make(chan packet, 64)
	go func() {
		defer close(ch)
		var buf [5]byte
		last := time.Now()
		for {
			if time.Since(last) > timeout {
				return
			}
			b, err := rb.ReadByte()
			if err == ringbuffer.ErrIsEmpty {
				runtime.Gosched()
				continue
			}
			if err != nil {
				return
			}
			switch st {
			case 0:
				if (1<<7)&b > 0 {
					buf[st] = b
					st++
				}
			default:
				if (1<<7)&b > 0 {
					st = 0
					continue
				}
				buf[st] = b
				if st == 4 {
					last = time.Now()
					ch <- packet{ts: last, buf: buf}
					st = 0
				} else {
					st++
					continue
				}
			}
		}
	}()
	return ch
}
