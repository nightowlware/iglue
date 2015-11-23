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

func TestFifoReadWrite(t *testing.T) {
	id := "baz"
	channel, _ := iglue.Register(id)
	defer iglue.Unregister(id)

	err := iglue.Send(id, "hello iglue\n")
	if err != nil {
		t.Errorf("Error encountered while Sending! - %s", err.Error())
	}

	fmt.Println("---------")
	fmt.Println("Attempting to receive from channel baz")
	fmt.Println("Received msg: ", <-channel)
}
