package main

import (
	"context"
	"errors"
	"log"

	"tinygo.org/x/bluetooth"
)

func main() {

	wait()

	adapter := bluetooth.DefaultAdapter

	try(adapter.Enable(), "unable to initialize bluetooth stack")
	log.Println("initialized bluetooth stack")

	m := mode()
	switch m {
	case "listen":
		listenMode(context.TODO(), adapter)
	default:
		try(errors.New("unknown mode: "+m), "invalid argument(s)")
	}

}
