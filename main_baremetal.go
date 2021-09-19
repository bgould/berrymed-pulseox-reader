// +build baremetal

package main

import (
	"log"
	"time"
)

func wait() {
	time.Sleep(3 * time.Second)
}

func address() string {
	return "00:A0:50:C8:E7:31"
}

func after(duration time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		/*
			for t := time.Now(); time.Since(t) < duration; {
				time.Sleep(1 * time.Millisecond)
			}
		*/
		println("about to sleep")
		time.Sleep(2 * time.Second)
		println("fired")
		ch <- time.Now()
		close(ch)
	}()
	println("scheduled")
	return ch
}

func mode() string {
	return "listen"
}

func try(err error, msg string) {
	if err != nil {
		for {
			log.Println(msg, ": ", err.Error())
			time.Sleep(2 * time.Second)
		}
	}
}
