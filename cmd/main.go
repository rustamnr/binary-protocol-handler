package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/rustamnr/binary-protocol-handler/internal/processor"
	"github.com/rustamnr/binary-protocol-handler/internal/protocol"
	"github.com/rustamnr/binary-protocol-handler/internal/worker"
)

const (
	workerCount   = 4
	taskQueueSize = 32
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to server")

	processor := &processor.DummyProcessor{}
	tasks := make(chan worker.MessageTask, taskQueueSize)

	// Запускаем worker-пул
	for range workerCount {
		go worker.Start(tasks, processor)
	}

	for {
		header, err := protocol.ReadHeader(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by server")
				break
			}
			log.Printf("Error reading header: %v", err)
			break
		}

		buf := make([]byte, header.DataLen)
		if _, err := io.ReadFull(conn, buf); err != nil {
			log.Printf("Error reading data: %v", err)
			break
		}

		tasks <- worker.MessageTask{
			Conn:   conn,
			Header: *header,
			Data:   buf,
		}
	}
}
