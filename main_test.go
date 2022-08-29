// package test
package main

import (
	"io"
	"net"
	"testing"
)

func TestAMQP(t *testing.T) {
	go (func() {
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:9663")
		s, _ := net.ListenTCP("tcp", addr)
		for {
			c, _ := s.AcceptTCP()
			io.Copy(c, c)
		}
	})()
	go main()
}
