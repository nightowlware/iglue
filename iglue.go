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
	MSG_SIZE_BYTES            = 1024
	CHANNEL_BUFFER_SIZE_ITEMS = 20480 // this many MSG_SIZE_BYTES worth of buffering
	PAYLOAD_SEPARATOR         = "|"
	FIFO_DIR                  = "/tmp/iglue"
	SHUTDOWN_HEADER           = "__SHUTDOWN__"
)

type Msg struct {
	Header  string // the header of the message: can be any string
	Payload string // the payload
}

// Returns a string representation of the IglueMsg
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
		return nil, err
	}

	// ensure machine-global fifo registery directory exists
	err = os.MkdirAll(FIFO_DIR, 0777)
	if err != nil {
		return nil, err
	}

	fifoPath, err := createFifo(idToFifoPath(iglueId))
	if err != nil {
		return nil, err
	}

	iglueChan := make(chan Msg, CHANNEL_BUFFER_SIZE_ITEMS)

	// launch a "receive" go-routine that continuously reads from the fifo
	// and stuffs the data into the channel:
	go func() {
		// blocks until writer opens fifo for writing
		fifo, err := os.OpenFile(fifoPath, os.O_RDONLY, 0600)

		if err != nil {
			fmt.Println("!!! OpenFile failed, goroutine exiting: ", fifoPath, "!!!")
		} else {

			buf := make([]byte, MSG_SIZE_BYTES)

			for {
				// blocks until data is available
				_, err := fifo.Read(buf)

				if err == nil {
					// read fixed-size messages, but trim off the
					// padded null bytes before pushing the string into the
					// channel
					msg := string(bytes.TrimRight(buf[:MSG_SIZE_BYTES], "\x00"))
					splits := strings.SplitN(msg, PAYLOAD_SEPARATOR, 2)

					// if we receive the special "shutdown" message, cleanup
					// and exit goroutine.
					if splits[0] == SHUTDOWN_HEADER {
						break
					}
					// otherwise, stuff Msg in the client's channel
					iglueChan <- Msg{splits[0], splits[1]}
				} else if err == io.EOF {
					// if we reach the end of the data,
					// simply skip this current iteration and go back
					// to blocking in the Read() call above
					continue
				} else {
					// quit on read error
					break
				}
			}
		}

		// cleanup
		close(iglueChan)
		return
	}()

	return iglueChan, err
}

func Unregister(iglueId string) error {
	// send a special shutdown message then delete the fifo
	Send(&Msg{SHUTDOWN_HEADER, ""}, iglueId)
	path := idToFifoPath(iglueId)
	return os.Remove(path)
}

// TODO: This function has a lot of potential for
// optimization.
func Send(igluemsg *Msg, iglueIdDest string) error {
	//fmt.Println("Send(): ")
	fifo, err := os.OpenFile(idToFifoPath(iglueIdDest), os.O_WRONLY, 0600)
	defer fifo.Close()

	if err != nil {
		return err
	}

	data := []byte(igluemsg.String())

	// check number of *bytes* (not characters) in data,
	// and make sure it's no more than the max allowed.
	padsize := MSG_SIZE_BYTES - len(data)
	if padsize < 0 {
		return errors.New("Send(): Message size exceeds max allowed!")
	}

	// msgbuf will be created with null bytes, which pad the message up
	// to MSG_SIZE_BYTES
	msgbuf := make([]byte, MSG_SIZE_BYTES)

	// copy message data into beginning of msgbuf
	for i, c := range data {
		msgbuf[i] = c
	}

	_, err = fifo.Write(msgbuf)
	//fmt.Println("Send(): After Write")
	return err
}

func idToFifoPath(iglueId string) string {
	return fmt.Sprintf("%s/%s", FIFO_DIR, iglueId)
}
