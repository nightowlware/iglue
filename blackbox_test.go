package iglue_test

import (
	"fmt"
	"github.com/nightowlware/iglue"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	iglue.Register("foo")
	iglue.Register("bar")

	// give some time for the underlying goroutine to
	// spin up and open the pipe
	time.Sleep(1000 * time.Millisecond)

	if err := iglue.Unregister("foo"); err != nil {
		t.Errorf("Could not unregister: %s", err.Error())
	}

	if err := iglue.Unregister("bar"); err != nil {
		t.Errorf("Could not unregister: %s", err.Error())
	}
}

func TestFifoRead(t *testing.T) {

	channel, _ := iglue.Register("baz")
	defer iglue.Unregister("baz")

	for {
		fmt.Println("---------")
		fmt.Println("Attempting to receive from channel baz")
		fmt.Println("!!! Received msg: ", <-channel)
	}
}
