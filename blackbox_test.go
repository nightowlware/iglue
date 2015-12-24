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
		t.Fatalf("Could not unregister: %s", err.Error())
	}

	if err := iglue.Unregister("bar"); err != nil {
		t.Fatalf("Could not unregister: %s", err.Error())
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
			t.Fatalf("Error encountered while Sending! - %s", err.Error())
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
		t.Fatalf("Sending to 'INVALID' succeeded when it shouldn't have!")
	}

	fmt.Println("Result of sending to 'INVALID' (should be an error): ", err)
}

func TestWriteStress(t *testing.T) {
	id := "stress"
	channel, err := iglue.Register(id)

	if err != nil {
		t.Fatalf("Registering for iglueId of < %s > failed!", id)
	}

	n := 40000

	pMsg := &iglue.Msg{"HEADER", "1234567890"}

	fmt.Printf("Sending %d messages over fifo:\n", n)
	for i := 1; i < n+1; i++ {
		iglue.Send(pMsg, "stress")
		if i%10000 == 0 {
			fmt.Println(i)
		}
		<-channel
		//fmt.Println(n, <-channel)
	}

	defer iglue.Unregister("stress")
}

func TestMultRegister(t *testing.T) {
	iglue.Register("test")
	if _, err := iglue.Register("test"); err == nil {
		t.Fatalf("Register called twice with no error!")
	}
	iglue.Unregister("test")
}

func TestMultUnregister(t *testing.T) {
	iglue.Register("test")
	iglue.Unregister("test")
	if iglue.Unregister("test") == nil {
		t.Fatalf("Unregister called twice with no error!")
	}
}

func TestStopRecv(t *testing.T) {
	channel, _ := iglue.Register("deleteme")
	iglue.Unregister("deleteme")
	time.Sleep(2000 * time.Millisecond)

	if _, alive := <-channel; alive {
		t.Fatalf("Recv Goroutine was not shutdown properly upon Unregister.")
	}
}

///////////////////////////////////////////////

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
