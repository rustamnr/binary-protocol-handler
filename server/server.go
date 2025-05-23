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
	fmt.Println("Сервер слушает на :5678")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка подключения:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	data := []byte("hello world")

	msg := makeMessage(1, 42, 777, data)
	_, err := conn.Write(msg)
	if err != nil {
		fmt.Println("Ошибка отправки:", err)
		return
	}
	fmt.Println("Сообщение отправлено клиенту")

	// Ожидаем ответ
	buf := make([]byte, 24)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}
	fmt.Printf("Получен ответ от клиента (%d байт): %x\n", n, buf[:n])
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
