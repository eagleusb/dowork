package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"git.sr.ht/~sircmpwn/dowork"
)

func main() {
	log.Println("Starting queue")
	log.Println("SIGINT to terminate")
	work.Start()

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
				log.Println("%v", err)
				continue
			}

			ntasks += 1
			work.Submit(func(n int) func(context.Context) error {
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
