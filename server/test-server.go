package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"net"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":5678")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server started on port 5678")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	messages := []struct {
		id    uint32
		param uint32
		data  []byte
	}{
		{42, 1, []byte("fast1")},
		{43, 2, []byte("slow1")},
		{44, 3, []byte("fast2")},
		{45, 4, []byte("slow2")},
	}

	for _, msg := range messages {
		packet := makeMessage(1, msg.id, msg.param, msg.data)
		conn.Write(packet)
		fmt.Printf("Message sent: ID=%d, Param=%d, Data=%s\n", msg.id, msg.param, msg.data)
	}

	// Читаем все ответы (24 байта × 4)
	for i := 0; i < len(messages); i++ {
		buf := make([]byte, 24)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}
		fmt.Printf("Received response: %d bytes\n", n)
	}
}

func makeMessage(msgType, id, param uint32, data []byte) []byte {
	header := new(bytes.Buffer)

	// Пишем 5 полей заголовка (всего 20 байт)
	binary.Write(header, binary.BigEndian, msgType)           // 0–3
	binary.Write(header, binary.BigEndian, id)                // 4–7
	binary.Write(header, binary.BigEndian, uint32(0))         // 8–11 (Response)
	binary.Write(header, binary.BigEndian, param)             // 12–15
	binary.Write(header, binary.BigEndian, uint32(len(data))) // 16–19

	headerBytes := header.Bytes()
	checksum := crc32.ChecksumIEEE(headerBytes)

	// Сформируем итоговый пакет
	final := new(bytes.Buffer)
	final.Write(headerBytes)
	binary.Write(final, binary.BigEndian, checksum) // 20–23
	final.Write(data)                               // Данные

	return final.Bytes()
}
