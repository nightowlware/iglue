package main

import (
	"fmt"
	"github.com/nightowlware/iglue"
)

func main() {
	fmt.Println("Process1 starting ...")

	_, err, statusptr := iglue.Register("p1")
	defer iglue.Unregister("p1")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 1000 && *statusptr == true; i++ {
		msg := iglue.Msg{"Header", fmt.Sprintf("Yo, this is p1, message # %d", i)}
		fmt.Println("P1: Sending to p2: ", msg.String())

		err := iglue.Send(&msg, "p2")
		if err != nil {
			panic(fmt.Errorf("Seems like p2 is not running yet; start p2 first then try again.\n%s", err))
		}
	}
}
