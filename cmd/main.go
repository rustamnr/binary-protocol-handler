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
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()
	fmt.Println("Подключение установлено")

	processor := &processor.DummyProcessor{}
	tasks := make(chan worker.MessageTask, taskQueueSize)

	// Запускаем worker-пул
	for i := 0; i < workerCount; i++ {
		go worker.Start(tasks, processor)
	}

	for {
		header, err := protocol.ReadHeader(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Соединение закрыто сервером")
				break
			}
			log.Printf("Ошибка чтения заголовка: %v", err)
			break
		}

		buf := make([]byte, header.DataLen)
		if _, err := io.ReadFull(conn, buf); err != nil {
			log.Printf("Ошибка чтения данных: %v", err)
			break
		}

		tasks <- worker.MessageTask{
			Conn:   conn,
			Header: *header,
			Data:   buf,
		}
	}
}
