package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/seka/playground-golang/server"
)

func main() {
	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		panic(err)
	}
	defer upg.Stop()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				log.Println("Upgrade failed:", err)
				continue
			}
			log.Println("Upgrade succeeded")
		}
	}()

	ln, err := upg.Fds.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln("Can't listen:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := server.NewEchoServer(server.EchoServerOption{
		ParentListener: ln,
	})
	go func() {
		s.Run(ctx)
	}()

	if err := upg.Ready(); err != nil {
		panic(err)
	}
	<-upg.Exit()

	time.AfterFunc(30*time.Second, func() {
		os.Exit(1)
	})
}
