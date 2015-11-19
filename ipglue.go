package ipglue

import (
	"fmt"
	"os"
	"io"
	"syscall"
)

const (
	MSG_SIZE_BYTES            = 512
	FIFO_DIR                  = "/tmp/ipglue"
	CHANNEL_BUFFER_SIZE_ITEMS = 20480 // 10MB max worth of buffering
)

func Register(name string) (<-chan string, error) {
	// ensure machine-global fifo registery directory exists
	err := os.MkdirAll(FIFO_DIR, 0777)
	if err != nil {
		panic(err)
	}

	fifoPath, err := createFifo(fmt.Sprintf("%s/%s", FIFO_DIR, name))
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

func Unregister(name string) error {
	path := fmt.Sprintf("%s/%s", FIFO_DIR, name)
	fmt.Println("Removing fifo", path)
	return os.Remove(path)
}

// createFifo currently only works on Linux.
// TODO: Implement for Windows/Mac.
func createFifo(path string) (string, error) {
	err := syscall.Mknod(path, syscall.S_IFIFO|0666, 0)
	if err != nil {
		return "FIFO_FAIL", fmt.Errorf("Could not create fifo: %s : %s", path, err.Error())
	}

	fmt.Println("Created fifo: ", path)
	return path, nil
}
