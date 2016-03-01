package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	var binary, binarySend, binaryReceive bool
	flag.BoolVar(&binary, "b", false, "use binary messages (overrides -bs and -br)")
	flag.BoolVar(&binarySend, "bs", false, "send binary messages")
	flag.BoolVar(&binaryReceive, "br", false, "receive binary messages")
	flag.Parse()
	rawurl := flag.Arg(0)

	// ws://echo.websocket.org:80
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatal("error parsing: ", err)
	}

	if binary {
		binaryReceive = true
		binarySend = true
	}

	conn, err := websocket.Dial(rawurl, "tcp", "http://"+u.Host)
	if err != nil {
		log.Fatal("error dialing: ", err)
	}

	go func(conn *websocket.Conn) {
		scanner := bufio.NewScanner(os.Stdin)
		if binarySend {
			scanner.Split(bufio.ScanBytes)
		}

		for scanner.Scan() {
			var toSend interface{}
			if binarySend {
				toSend = scanner.Bytes()
			} else {
				toSend = scanner.Text()
			}
			if err := websocket.Message.Send(conn, toSend); err != nil {
				log.Fatalf("error writing message: %s", err)
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Fatalf("error scanning stdin: %s", err)
		}
		// We are done, force exit in case the receiver is blocked
		os.Exit(0)
	}(conn)

	var msg string
	var binMsg []byte
	var toReceive interface{} = &msg
	if binaryReceive {
		toReceive = &binMsg
	}

	for {
		err := websocket.Message.Receive(conn, toReceive)
		if err != nil {
			if err != io.EOF {
				log.Fatal("error recv: ", err)
			}
			break
		}
		if binaryReceive {
			os.Stdout.Write(binMsg)
		} else {
			fmt.Println(msg)
		}
	}
}
