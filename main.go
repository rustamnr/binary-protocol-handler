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

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	fmt.Println("Подключение установлено")

	processor := &DummyProcessor{}

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

		data := make([]byte, header.DataLen)
		if _, err := io.ReadFull(conn, data); err != nil {
			log.Printf("Ошибка чтения данных: %v", err)
			break
		}

		response := processor.ProcessMessage(header.Param, data)

		// Ответ без данных, только заголовок
		respHeader := Header{
			Type:     header.Type,
			ID:       header.ID,
			Response: uint32(response),
			Param:    0,
			DataLen:  0,
		}
		if err := WriteHeader(conn, respHeader); err != nil {
			log.Printf("Ошибка отправки ответа: %v", err)
			break
		}

		fmt.Printf("Обработано сообщение ID=%d, ответ=%d\n", header.ID, response)
	}
}

type MessageProcessor interface {
	ProcessMessage(param uint32, data []byte) uint16
}

type DummyProcessor struct{}

func (p *DummyProcessor) ProcessMessage(param uint32, data []byte) uint16 {
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	return sum
}

const headerSize = 24

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
	if err := binary.Write(buf, binary.BigEndian, checksum); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}
