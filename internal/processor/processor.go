package processor

import (
	"io"
	"log"
)

type MessageProcessor interface {
	ProcessMessage(param uint32, data io.Reader) uint16
}

type DummyProcessor struct{}

func (p *DummyProcessor) ProcessMessage(param uint32, data io.Reader) uint16 {
	var (
		buf = make([]byte, 4096)
		sum uint16
	)

	for {
		n, err := data.Read(buf)
		for i := 0; i < n; i++ {
			sum += uint16(buf[i])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading data: %v", err)
			break
		}
	}

	return sum
}
