// +build !baremetal

package main

import (
	"log"
)

func wait() {

}

func init() {
	address = "00:A0:50:C8:E7:31" //os.Args[1]
}

func try(err error, msg string) {
	if err == nil {
		return
	}
	log.Fatal(msg, ": ", err.Error())
}
