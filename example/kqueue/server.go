package main

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/Softwarekang/knet"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatal(err)
	}

	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		log.Fatal(err)
	}

	poller := knet.NewDefaultPoller()
	listenerFD := int(file.Fd())
	onRead := func() error {
		nfd, stockade, err := syscall.Accept(listenerFD)
		if err != nil {
			log.Fatal(err)
		}

		stockadeInt4 := stockade.(*syscall.SockaddrInet4)
		tcpAddr := &net.TCPAddr{
			IP:   stockadeInt4.Addr[0:],
			Port: stockadeInt4.Port,
		}

		fmt.Printf("server %s get accept new client conn:%v \n", listener.Addr().String(), tcpAddr.String())
		if err := poller.Register(&knet.NetFileDesc{
			FD: nfd,
			NetPollListener: knet.NetPollListener{
				OnRead: func() error {
					buf := make([]byte, 4)
					n, err := syscall.Read(nfd, buf)
					if err != nil {
						return err
					}

					fmt.Printf("server %s read %d bytes data from  client:%s, data:%s\n", tcpAddr.String(), n, tcpAddr.String(), string(buf))
					return nil
				}, OnInterrupt: func() error {
					fmt.Printf("client conn %s closed\n", tcpAddr.String())
					return poller.Register(&knet.NetFileDesc{
						FD: nfd,
					}, knet.DeleteRead)
				},
			},
		}, knet.Read); err != nil {
			return err
		}
		return nil
	}

	if err = poller.Register(&knet.NetFileDesc{
		FD: listenerFD,
		NetPollListener: knet.NetPollListener{
			OnRead: onRead,
		},
	}, knet.Read); err != nil {
		log.Fatal(err)
	}

	// block
	poller.Wait()
}