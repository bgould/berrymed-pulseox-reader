// +build !baremetal

package main

import (
	"log"
	"time"
)

func wait() {

}

func mode() string {
	return "listen"
}

func address() string {
	return "00:A0:50:C8:E7:31" //os.Args[1]
}

func after(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

func try(err error, msg string) {
	if err == nil {
		return
	}
	log.Fatal(msg, ": ", err.Error())
}
