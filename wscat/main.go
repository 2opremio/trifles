package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"golang.org/x/net/websocket"
)

func receiveLoop(binary bool, conn *websocket.Conn, stop chan int) {
	var (
		msg       string
		binMsg    []byte
		toReceive interface{} = &msg
	)

	if binary {
		toReceive = &binMsg
	}

	for {
		if err := websocket.Message.Receive(conn, toReceive); err != nil {
			status := 0
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "error receiving:", err)
				status = 1
			}
			stop <- status
			return
		}
		if binary {
			os.Stdout.Write(binMsg)
		} else {
			fmt.Println(msg)
		}
	}
	stop <- 0
}

func sendLoop(binary bool, conn *websocket.Conn, stop chan int) {
	scanner := bufio.NewScanner(os.Stdin)
	if binary {
		scanner.Split(bufio.ScanBytes)
	}

	for scanner.Scan() {
		var toSend interface{}
		if binary {
			toSend = scanner.Bytes()
		} else {
			toSend = scanner.Text()
		}
		if err := websocket.Message.Send(conn, toSend); err != nil {
			fmt.Fprintln(os.Stderr, "error sending:", err)
			stop <- 1
			return
		}
	}
	status := 0
	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, "error reading:", err)
		status = 1
	}
	stop <- status
}

func main() {
	var (
		binary, binarySend, binaryReceive bool
		stop                              = make(chan int)
	)
	flag.BoolVar(&binary, "b", false, "use binary messages (overrides -bs and -br)")
	flag.BoolVar(&binarySend, "bs", false, "send binary messages")
	flag.BoolVar(&binaryReceive, "br", false, "receive binary messages")
	flag.Parse()
	rawurl := flag.Arg(0)

	u, err := url.Parse(rawurl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing url:", err)
		os.Exit(1)
	}

	if binary {
		binaryReceive = true
		binarySend = true
	}

	conn, err := websocket.Dial(rawurl, "tcp", "http://"+u.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error dialing:", err)
		os.Exit(1)
	}

	go sendLoop(binarySend, conn, stop)
	go receiveLoop(binaryReceive, conn, stop)
	status := <-stop
	os.Exit(status)
}
