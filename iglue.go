package iglue

import (
	"fmt"
	"io"
	"os"
)

const (
	MSG_SIZE_BYTES            = 1024
	FIFO_DIR                  = "/tmp/iglue"
	CHANNEL_BUFFER_SIZE_ITEMS = 20480 // this many MSG_SIZE_BYTES worth of buffering
)

func Register(iglueId string) (<-chan string, error) {
	// ensure machine-global fifo registery directory exists
	err := os.MkdirAll(FIFO_DIR, 0777)
	if err != nil {
		panic(err)
	}

	fifoPath, err := createFifo(idToFifoPath(iglueId))
	if err != nil {
		panic(err)
	}

	fifoChan := make(chan string, CHANNEL_BUFFER_SIZE_ITEMS)

	// launch a go-routine that continuously reads from the fifo
	// and stuffs the data into the channel:
	go func() {
		// Note: os.Open blocks until another process
		// writes into fifo!
		fifo, err := os.Open(fifoPath)
		if err != nil {
			fmt.Println("!!! Fifo was removed, channel will never return data:", fifoPath, "!!!")
		}

		buf := make([]byte, MSG_SIZE_BYTES)

		for {
			n, err := fifo.Read(buf)
			if err == nil {
				msg := string(buf[:n])
				fifoChan <- msg
			} else if err == io.EOF {
				// not intuitive: we have to re-open the fifo if we ever get a
				// read of zero bytes (EOF), so that we block again. Ugly,
				// but those are the semantics of unix pipes.
				fifo, err = os.Open(fifoPath)
			} else {
				fmt.Println(err)
				// quit on read error
				return
			}
		}
	}()

	return fifoChan, err
}

func Unregister(iglueId string) error {
	path := idToFifoPath(iglueId)
	fmt.Println("Removing fifo", path)
	return os.Remove(path)
}

// TODO: This function has a lot of potential for 
// optimization.
func Send(iglueId string, msg string) error {
	fifo, err := os.OpenFile(idToFifoPath(iglueId), os.O_APPEND|os.O_WRONLY, 0600)
	defer fifo.Close()

	if err != nil {
		return err
	}

	_, err = fifo.WriteString(msg)
	return err
}

func idToFifoPath(iglueId string) string {
	return fmt.Sprintf("%s/%s", FIFO_DIR, iglueId)
}
