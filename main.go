package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	pbstream "github.com/hashicorp/nomad/nomad/stream/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// TLS setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// these certs were copied from nomad/dev/tls_cluster/certs
	certificate, err := tls.LoadX509KeyPair(
		"server.pem",
		"server-key.pem",
	)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile("nomad-ca.pem")
	if err != nil {
		log.Fatalf("failed to read ca cert: %s", err)
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatal("failed to append certs")
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost",
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	// gRPC dial Nomad's HTTP endpoint
	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:4646",
		grpc.WithTransportCredentials(transportCreds),
	)
	if err != nil {
		log.Fatalf("error dialing: %v", err)
	}

	// create EventStreamClient
	client := pbstream.NewEventStreamClient(conn)

	ctx, cancelSub := context.WithCancel(context.Background())

	// subscribe to the event stream
	sub, err := client.Subscribe(ctx, &pbstream.SubscribeRequest{
		Index: 0,
		Topics: []*pbstream.TopicFilter{
			{
				// Receive all available events
				Topic: pbstream.Topic_All,

				// Receive only Deployment-related events
				// Topic: pbstream.Topic_Deployment,

				// Receive only Job-related events that match the FilterKeys (the job ID/name)
				// Topic:      pbstream.Topic_Job,
				// FilterKeys: []string{"my-job"},
			},
		},
	})
	if err != nil {
		log.Fatalf("error subscribing: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)

	go func() {
		<-sigCh
		log.Printf("received signal, exiting")
		cancelSub()
		cancel()
		os.Exit(0)
	}()

	for {
		eventBatch, err := sub.Recv()
		if err != nil {
			log.Printf("error receiving: %v\n", err)
			break
		}

		log.Printf("======= batch %d =======\n", eventBatch.Index)

		for _, event := range eventBatch.Event {
			log.Printf("Type=%s Topic=%s Payload=%v\n", event.Type, event.Topic, event.Payload)
		}
	}
}
