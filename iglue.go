package iglue

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	PAYLOAD_SEPARATOR         = "|"
	MSG_SIZE_BYTES            = 1024
	FIFO_DIR                  = "/tmp/iglue"
	CHANNEL_BUFFER_SIZE_ITEMS = 20480 // this many MSG_SIZE_BYTES worth of buffering
)

type Msg struct {
	Header  string // the header of the message: can be any string
	Payload string // the payload
}

// String returns a string representation of the IglueMsg
func (m *Msg) String() string {
	return fmt.Sprint(m.Header, PAYLOAD_SEPARATOR, m.Payload)
}

func validateIglueId(iglueId string) error {
	if strings.Contains(iglueId, PAYLOAD_SEPARATOR) {
		return fmt.Errorf("Iglue ID cannot contain the Payload Separator character <%s>", PAYLOAD_SEPARATOR)
	}

	return nil
}

func Register(iglueId string) (<-chan Msg, error) {
	err := validateIglueId(iglueId)
	if err != nil {
		panic(err)
	}

	// ensure machine-global fifo registery directory exists
	err = os.MkdirAll(FIFO_DIR, 0777)
	if err != nil {
		panic(err)
	}

	fifoPath, err := createFifo(idToFifoPath(iglueId))
	if err != nil {
		panic(err)
	}

	iglueChan := make(chan Msg, CHANNEL_BUFFER_SIZE_ITEMS)

	// launch a go-routine that continuously reads from the fifo
	// and stuffs the data into the channel:
	go func() {
		// Note: we have to open the file as read-write so as to avoid
		// the blocking "feature" of the open call when it's done in read or write
		// only mode.
		fifo, err := os.OpenFile(fifoPath, os.O_RDWR, 0600)
		if err != nil {
			fmt.Println("!!! Fifo was removed, channel will never return data:", fifoPath, "!!!")
		}

		buf := make([]byte, MSG_SIZE_BYTES)

		for {
			// Blocks until data is available
			_, err := fifo.Read(buf)

			if err == nil {
				// read fixed-size messages, but trim off the
				// padded null bytes before pushing the string into the
				// channel
				msg := string(bytes.TrimRight(buf[:MSG_SIZE_BYTES], "\x00"))
				splits := strings.SplitN(msg, PAYLOAD_SEPARATOR, 2)
				iglueChan <- Msg{splits[0], splits[1]}
			} else if err == io.EOF {
				// if we reach the end of the data,
				// simply skip this current iteration and go back
				// to blocking in the Read() call above
				continue
			} else {
				fmt.Println("!!! ERROR: ", err)
				// quit on read error
				return
			}
		}
	}()

	return iglueChan, err
}

func Unregister(iglueId string) error {
	path := idToFifoPath(iglueId)
	//fmt.Println("Removing fifo", path)
	return os.Remove(path)
}

// TODO: This function has a lot of potential for
// optimization.
func Send(igluemsg *Msg, iglueIdDest string) error {
	fifo, err := os.OpenFile(idToFifoPath(iglueIdDest), os.O_WRONLY, 0600)
	defer fifo.Close()

	if err != nil {
		return err
	}

	msgbuf := []byte(igluemsg.String())

	// check number of *bytes* (not characters) in msg,
	// and make sure it's no more than the max allowed.
	padsize := MSG_SIZE_BYTES - len(msgbuf)
	if padsize < 0 {
		return errors.New("Send(): Message size exceeds max allowed!")
	}

	// force sending fixed-size messages by padding up to MSG_SIZE_BYTES
	// the pad is made of null-bytes
	_, err = fifo.Write(append(msgbuf, make([]byte, padsize)...))
	return err
}

func idToFifoPath(iglueId string) string {
	return fmt.Sprintf("%s/%s", FIFO_DIR, iglueId)
}
