package main

import (
	"fmt"
	"github.com/nightowlware/iglue"
)

func main() {
	fmt.Println("Process2 starting, waiting for Process1 ...")

	channel, err, statusptr := iglue.Register("p2")
	defer iglue.Unregister("p2")

	if err != nil {
		fmt.Println("Could not Register p2 !!!")
		panic(err)
	}

	for *statusptr == true {
		msg := <-channel
		fmt.Println("P2: Received from p1: ", msg.String())
	}
}
