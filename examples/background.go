package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"git.sr.ht/~sircmpwn/dowork"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting queue")
	log.Println("SIGINT to terminate")
	work.Start()

	mux := &http.ServeMux{}
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Handler: mux}
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	log.Printf("Prometheus listening on :%d", listen.Addr().(*net.TCPAddr).Port)
	go server.Serve(listen)

	ntasks := 0
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			log.Println("Enter time.ParseDuration string:")

			input, _ := reader.ReadString('\n')
			d, err := time.ParseDuration(strings.Trim(input, "\n"))
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			ntasks += 1
			work.Submit(func(n int) work.TaskFunc {
				return func(ctx context.Context) error {
					log.Printf("Task %d started", n)
					select {
					case <-time.After(d):
						break
					case <-ctx.Done():
						break
					}
					log.Printf("Task %d complete", n)
					return nil
				}
			}(ntasks))
		}
	}()

	<-sig
	log.Println("Shutting down queue...")
	work.Shutdown()
	log.Println("Shut down.")
}
