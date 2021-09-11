// +build !baremetal

package main

import "os"

func wait() {

}

func init() {
	address = os.Args[1]
}
