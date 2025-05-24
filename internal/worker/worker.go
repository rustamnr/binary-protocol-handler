package worker

import (
	"bytes"
	"log"
	"net"

	"github.com/rustamnr/binary-protocol-handler/internal/processor"
	"github.com/rustamnr/binary-protocol-handler/internal/protocol"
)

type MessageTask struct {
	Conn   net.Conn
	Header protocol.Header
	Data   []byte
}

func Start(tasks <-chan MessageTask, proc processor.MessageProcessor) {
	for task := range tasks {
		r := bytes.NewReader(task.Data)
		response := proc.ProcessMessage(task.Header.Param, r)

		respHeader := protocol.Header{
			Type:     task.Header.Type,
			ID:       task.Header.ID,
			Response: uint32(response),
			Param:    0,
			DataLen:  0,
		}

		if err := protocol.WriteHeader(task.Conn, respHeader); err != nil {
			log.Printf("Error writing response header for ID=%d: %v", task.Header.ID, err)
			continue
		}

		log.Printf("Processed message ID=%d, response=%d", task.Header.ID, response)
	}
}
