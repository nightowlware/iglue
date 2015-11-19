package ipglue_test

import (
	"github.com/nightowlware/ipglue"
	"fmt"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	ipglue.Register("foo")
	ipglue.Register("bar")

	// give some time for the underlying goroutine to 
	// spin up and open the pipe
	time.Sleep(1000 * time.Millisecond)

	if err := ipglue.Unregister("foo"); err != nil {
		t.Errorf("Could not unregister: %s", err.Error())
	}

	if err := ipglue.Unregister("bar"); err != nil {
		t.Errorf("Could not unregister: %s", err.Error())
	}
}

func TestFifoRead(t *testing.T) {

	channel, _ := ipglue.Register("baz")
	defer ipglue.Unregister("baz")

	for {
		fmt.Println("---------")
		fmt.Println("Attempting to receive from channel baz")
		fmt.Println("!!! Received msg: ", <-channel)
	}
}
