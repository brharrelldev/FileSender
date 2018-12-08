package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var sendFile string
var recvFile string

type message struct {
	Role string `json:"role"`
	Data []byte `json:"data"`

}

func main()  {


	rootCmd := &cobra.Command{
		Use: "fsender",
		Short: "file sending utility",

	}

	sendCmd :=  &cobra.Command{
		Use: "send",
		Short: "file to send",
		Run: func(cmd *cobra.Command, args []string) {

			m := message{}

			f, err := os.Open(sendFile)
			if err != nil{
				logrus.Fatalf("could not open file %s", sendFile)
			}

			defer f.Close()

			fStats, err := f.Stat()
			if err != nil{
				logrus.Fatalf("could not open file %s", sendFile)
			}

			fileSize := fStats.Size()

			if fileSize > 4096{
				log.Println("file is to large to be sent, exit")
				os.Exit(1)
			}

			sFile, err := ioutil.ReadFile(sendFile)
			if err != nil{
				panic(err)
			}

			conn, err := net.Dial("tcp", ":3000")
			if err != nil{
				log.Fatalf("could not connect to host %v\n", err)
			}

			m.Role = "sender"
			m.Data = sFile


			jsonMessage, err := json.Marshal(m)


			if err != nil{
				logrus.Errorf("could not serialize json %v\n", err)
				conn.Close()
			}

			fmt.Println(jsonMessage)

			_, err = conn.Write(jsonMessage)
			if err != nil{
				conn.Close()
			}


		},

	}

	recv := &cobra.Command{
		Use: "recv",
		Short: "receive file from server",
		Run: func(cmd *cobra.Command, args []string) {

			conn, err := net.Dial("tcp", ":3000")

			defer conn.Close()
			if err != nil{
				conn.Close()
				log.Fatalf("could not connect to server %v\n", err)
			}

			//buf := make([]byte, 0, 4096)
			read, err := ioutil.ReadAll(conn)
			if err != nil{
				if err != io.EOF{
					logrus.Errorf("could not read due to err %v\n", err)
				}
				logrus.Errorf("could not read information from connection %v\n", err)
				conn.Close()
			}


			f, err := os.Create("new_file.txt")

			defer f.Close()
			if err != nil{
				panic(err)
			}

			f.Write(read)


		},
	}

	sendCmd.Flags().StringVar(&sendFile, "sendfile", "file.txt",  "path of text file to send" )


	rootCmd.AddCommand(sendCmd, recv)
	rootCmd.Execute()

}
