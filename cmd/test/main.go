package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8088", time.Second*30)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			fmt.Println("done")
		}()
		for {
			data := make([]byte, 10)
			n, err := conn.Read(data)
			fmt.Println("recv info: ", n, err, data)
			if err != nil {
				fmt.Println("not nil")
				break
			}
		}

	}()

	data := make([]byte, 10)
	binary.BigEndian.PutUint16(data, 8)
	data[2] = 0x1
	data[3] = 0x2
	data[4] = 0x3
	data[5] = 0x4
	data[6] = 0x5
	data[7] = 0x6
	data[8] = 0x7
	data[9] = 0x8
	conn.Write(data)

	time.Sleep(time.Second * 10)
	conn.Write(data)
	time.Sleep(time.Second * 60)
	conn.Write(data)
	wg.Wait()
	fmt.Println("exit")
}
