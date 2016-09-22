package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/drone/mq/stomp"

	"github.com/urfave/cli"
)

// we provide very simple and naive benchmarking. This is used to profile
// the client and server under load. This should not be used for official
// performance benchmarks or product benchmark comparisions.

var comandBench = cli.Command{
	Name:  "bench",
	Usage: "tools for profiling and benchmarking",
	Subcommands: []cli.Command{
		{
			Name:   "pub",
			Usage:  "benchmark publish",
			Action: benchPub,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "topic",
					Usage: "name of the topic",
					Value: "/topic/bench.%d",
				},
				cli.IntFlag{
					Name:  "message-count",
					Usage: "number of messages to send",
					Value: 100000,
				},
				cli.IntFlag{
					Name:  "client-count",
					Usage: "number of client connections to use",
					Value: 1,
				},
			},
		},
		{
			Name:   "pubsub",
			Usage:  "benchmark publish and subscribe",
			Action: bench,
			Before: setup,
			After:  teardown,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "topic",
					Usage: "name of the topic",
					Value: "/topic/bench",
				},
				cli.IntFlag{
					Name:  "message-count",
					Usage: "number of messages to send",
					Value: 100000,
				},
			},
		},
	},
}

// executes a combined publish / subscribe benchmark using a single client
// connection. the benchmark blocks until all messages are published to the
// server and subsequently forwarded to, and processed by, the subscriber.
func bench(c *cli.Context) error {
	fmt.Println("Performing Publish/Subscribe performance test")

	var (
		wg sync.WaitGroup

		messages = c.Int("message-count")
		topic    = c.String("topic")
	)

	handler := func(m *stomp.Message) {
		wg.Done()
		m.Release()
	}

	_, err := client.Subscribe(topic, stomp.HandlerFunc(handler))
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	wg.Add(messages)

	for i := 0; i < messages; i++ {
		err = client.Send(topic, []byte("ping"))
		if err != nil {
			log.Fatal(err)
		}
	}

	wg.Wait()

	elapsed := time.Now().Sub(start)
	fmt.Printf(resultf, 1, elapsed,
		float64(messages)/elapsed.Seconds(),
	)

	return nil
}

// executes a publish-only benchmark using one or many client connections.
// the benchmark blocks until all messages are dispatched by the client.
//
// Note this does not guarantee the server has received or processed the
// messages as they could still be buffered by the client connection. This
// benmcharmk therefore still requires some improvement.
func benchPub(c *cli.Context) error {
	fmt.Println("Performing Publish performance test")

	var (
		wg sync.WaitGroup

		messages = c.Int("message-count")
		count    = c.Int("client-count")
		topic    = c.String("topic")

		clients = make([]*stomp.Client, count)
	)

	// initialize N client connections
	for i := range clients {
		var err error
		clients[i], err = createClient(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	// batch defines a function for sending the specified number of
	// messages in batch using the specified client.
	batch := func(client *stomp.Client, topic string, messages int) (err error) {
		for i := 0; i < messages; i++ {
			err = client.Send(topic, []byte("ping"))
			if err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	start := time.Now()
	wg.Add(count)

	// distribute messages across the group of clients
	for i, client := range clients {
		go func(client *stomp.Client, num int) {
			batch(client, fmt.Sprintf(topic, num), messages/count)
			wg.Done()
		}(client, i)
	}

	wg.Wait()

	elapsed := time.Now().Sub(start)
	fmt.Printf(resultf, count, elapsed,
		float64(messages)/elapsed.Seconds(),
	)

	return nil
}

var resultf = `
clients: %d
elapsed: %s
msg/sec: %.2f
latency: n/a

`
