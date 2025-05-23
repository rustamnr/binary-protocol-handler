package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
)

const (
	headerSize    = 24
	workerCount   = 4
	taskQueueSize = 32
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	fmt.Println("Подключение установлено")

	processor := &DummyProcessor{}
	taskChan := make(chan MessageTask, taskQueueSize)

	// Запускаем воркеры
	for i := 0; i < workerCount; i++ {
		go worker(taskChan, processor)
	}

	for {
		fmt.Println("Чтение заголовка...")
		header, err := ReadHeader(conn)
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
		dataReader := bytes.NewReader(buf)

		task := MessageTask{
			Conn:   conn,
			Header: *header,
			Data:   dataReader,
		}
		taskChan <- task
	}
}

// ---------------- Worker-пул ----------------

type MessageTask struct {
	Conn   net.Conn
	Header Header
	Data   io.Reader
}

func worker(tasks <-chan MessageTask, processor MessageProcessor) {
	for task := range tasks {
		response := processor.ProcessMessage(task.Header.Param, task.Data)

		respHeader := Header{
			Type:     task.Header.Type,
			ID:       task.Header.ID,
			Response: uint32(response),
			Param:    0,
			DataLen:  0,
		}

		if err := WriteHeader(task.Conn, respHeader); err != nil {
			log.Printf("Ошибка отправки ответа ID=%d: %v", task.Header.ID, err)
			continue
		}

		fmt.Printf("Обработано сообщение ID=%d, ответ=%d\n", task.Header.ID, response)
	}
}

// ---------------- Протокол ----------------

type Header struct {
	Type     uint32
	ID       uint32
	Response uint32
	Param    uint32
	DataLen  uint32
	Checksum uint32
}

func ReadHeader(r io.Reader) (*Header, error) {
	buf := make([]byte, headerSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	fmt.Printf("RAW HEADER: %x\n", buf)

	h := &Header{
		Type:     binary.BigEndian.Uint32(buf[0:4]),
		ID:       binary.BigEndian.Uint32(buf[4:8]),
		Response: binary.BigEndian.Uint32(buf[8:12]),
		Param:    binary.BigEndian.Uint32(buf[12:16]),
		DataLen:  binary.BigEndian.Uint32(buf[16:20]),
		Checksum: binary.BigEndian.Uint32(buf[20:24]),
	}

	expectedChecksum := crc32.ChecksumIEEE(buf[:20])
	if expectedChecksum != h.Checksum {
		return nil, fmt.Errorf("ошибка контрольной суммы: expected %08x, got %08x", expectedChecksum, h.Checksum)
	}

	return h, nil
}

func WriteHeader(w io.Writer, h Header) error {
	buf := new(bytes.Buffer)
	fields := []uint32{h.Type, h.ID, h.Response, h.Param, h.DataLen}
	for _, v := range fields {
		if err := binary.Write(buf, binary.BigEndian, v); err != nil {
			return err
		}
	}
	checksum := crc32.ChecksumIEEE(buf.Bytes())
	binary.Write(buf, binary.BigEndian, checksum)
	_, err := w.Write(buf.Bytes())
	return err
}

// ---------------- Обработчик ----------------

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
			log.Printf("Ошибка чтения данных: %v", err)
			break
		}
	}

	return sum
}
