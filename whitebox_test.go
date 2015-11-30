package iglue

import (
	_ "fmt"
	_"os"
	"testing"
	"time"
)

func TestStopRecv(t *testing.T) {
	_, _, status := Register("deleteme")

	Send(&Msg{"Header", "useless message"}, "deleteme")

	Unregister("deleteme")
	time.Sleep(2000 * time.Millisecond)

	if *status == true {
		t.Errorf("Recv Goroutine was not shutdown properly upon Unregister.")
	}
}
