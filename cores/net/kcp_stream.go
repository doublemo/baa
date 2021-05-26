// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package net

import (
	"net"

	"github.com/golang/snappy"
)

type KCPStream struct {
	conn net.Conn
	w    *snappy.Writer
	r    *snappy.Reader
}

func (c *KCPStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *KCPStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	err = c.w.Flush()
	return n, err
}

func (c *KCPStream) Close() error {
	return c.conn.Close()
}

func NewKCPStream(conn net.Conn) *KCPStream {
	c := new(KCPStream)
	c.conn = conn
	c.w = snappy.NewBufferedWriter(conn)
	c.r = snappy.NewReader(conn)
	return c
}
