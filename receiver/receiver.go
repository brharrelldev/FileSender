package main

import (
	"fmt"
	"log"
	"net"
)

func main()  {

	conn, err := net.Dial("tcp", ":3000")

	if err != nil{
		log.Println(err)
	}

	buf := make([]byte, 4096)

	read, err := conn.Read(buf)

	fmt.Println(buf[:read])








}
