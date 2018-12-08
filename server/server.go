package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type relay struct {
	addr       string
	port       string
	message    []byte
	outmessage []byte
	wg         sync.WaitGroup
	mu         sync.Mutex
}

type receivedMsg struct {
	Role string `json:"role"`
	Data []byte `json:"data"`
}

func (r *relay) listen() {

	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("could not allocate listener %s\n", err)
	}
	defer lis.Close()

	for {

		var msg receivedMsg
		writeChan := make(chan []byte,4096)
		readChan := make(chan []byte, 1)
		errChan := make(chan error,1)
		conn, err := lis.Accept()
		if err != nil {
			conn.Close()
		}

		serverFile, err := filepath.Glob("tmp/*")
		if err != nil {
			logrus.Fatalf("could not read path %v\n", err)

		}

		if len(serverFile) <= 0 {

			m, err := ioutil.ReadAll(conn)
			if err != nil {
				conn.Close()
			}

			if err := json.Unmarshal(m, &msg); err != nil {
				conn.Close()
			}

			switch msg.Role {
			case "sender":
				r.message = msg.Data

				r.wg.Add(1)
				defer r.wg.Done()

				go r.readInbound(conn, readChan, errChan)
				r.wg.Wait()
				select {
				case r := <-readChan:
					if _, err := os.Stat("tmp"); os.IsNotExist(err) {
						if err := os.Mkdir("tmp", os.ModePerm); err != nil {
							errChan <- err
							log.Fatalf("error creating directory %v\n", err)
						}
					}
					if err := ioutil.WriteFile("tmp/tempfile.txt", r, os.ModePerm); err != nil {
						log.Fatalf("error writing file to disk %v\n", err)

					}

					close(readChan)
					conn.Close()
				case e := <-errChan:
					fmt.Println("error received ", e)
				default:
					fmt.Println("no message received!")

				}

			default:
				fmt.Println("unknown error received")
			}

		} else {
			r.wg.Add(1)
			go r.writeOutbound(conn, writeChan, errChan)
			r.wg.Wait()
			select {
			case w := <-writeChan:
				fmt.Println(w)
				_, err := conn.Write(w)
				if err != nil {
					logrus.Errorf("could not write %v\n", err)
				}


				conn.Close()


			}

		}

	}

}

func (re *relay) readInbound(conn net.Conn, receive chan []byte, receiveErr chan error) {


	if len(re.message) <= 0 {
		fmt.Println("zero length file detected")
		receiveErr <- errors.New("zero byte file received")
		re.wg.Done()
		return
	}

	receive <- re.message

	conn.Close()
	re.wg.Done()

}

func (re *relay) writeOutbound(conn net.Conn, send chan []byte, writeErr chan error) {

	w, err := ioutil.ReadFile("tmp/tempfile.txt")


	if err != nil {
		writeErr <- err
		logrus.Errorf("could not read file %v\n", err)
		conn.Close()
		re.wg.Done()

	}

	send <- w
	re.wg.Done()

}

func StartServer(address, port string) {

	srv := relay{
		addr: address,
		port: port,
	}

	srv.listen()

}

func main() {

	StartServer("localhost", "3000")

}
