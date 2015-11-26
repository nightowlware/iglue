package iglue_test

import (
	"fmt"
	"github.com/nightowlware/iglue"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// setup a signal handler so that "Ctrl-\" shows a goroutine-trace.
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		buf := make([]byte, 1<<20)
		for {
			<-sigs
			runtime.Stack(buf, true)
			fmt.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)
		}
	}()

	os.Exit(m.Run())
}

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

	msgs := []string{"hello iglue", "this is iglue ipc", "and it works great!", "quit"}

	for _, msg := range msgs {
		err := iglue.Send(&iglue.Msg{"HEADER", msg}, "baz")
		if err != nil {
			t.Errorf("Error encountered while Sending! - %s", err.Error())
			t.FailNow()
		}
	}

	recvmsg := &iglue.Msg{}
	for recvmsg.Payload != "quit" {
		fmt.Println("Attempting to receive from channel baz")
		*recvmsg = <-channel
		fmt.Println("Received msg: ", recvmsg)
	}
}

func TestSendBadDest(t *testing.T) {
	err := iglue.Send(&iglue.Msg{"head", "tail"}, "INVALID")
	if err == nil {
		t.Errorf("Sending to 'INVALID' succeeded when it shouldn't have!")
	}

	fmt.Println("Result of sending to 'INVALID': ", err)
}

func TestWriteStress(t *testing.T) {
	channel, _ := iglue.Register("stress")

	pMsg := &iglue.Msg{"HEADER", "1234567890"}

	for n := 0; n < 20000; n++ {
		iglue.Send(pMsg, "stress")
		<-channel
		//fmt.Println(n, <-channel)
	}

	defer iglue.Unregister("stress")
}

func BenchmarkThroughput(b *testing.B) {
	channel, _ := iglue.Register("benchmark")
	pMsg := &iglue.Msg{"HEADER", "Benchmark message"}

	for n := 0; n < b.N; n++ {
		err := iglue.Send(pMsg, "benchmark")
		if err != nil {
			fmt.Println(err)
		}
		<-channel
	}

	defer iglue.Unregister("benchmark")
}
