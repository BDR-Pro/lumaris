package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"
	nakama "github.com/heroiclabs/nakama-go"
)

// JobRequest defines the structure of a compute job
type JobRequest struct {
	Image   string `json:"image"`    // Docker image to use
	Command string `json:"command"`  // Command to run inside the container
	BuyerID string `json:"buyer_id"` // ID of the buyer requesting the job
	JobID   string `json:"job_id"`   // Unique identifier for the job
}

func main() {
	// Parse command line flags
	nakamaServer := flag.String("server", "127.0.0.1:7350", "Nakama server address")
	httpKey := flag.String("key", "defaultkey", "Nakama server HTTP key")
	useSSL := flag.Bool("ssl", false, "Use SSL for connection")
	image := flag.String("image", "python:3.10", "Docker image to use")
	command := flag.String("command", "python -c 'print(\"Hello from compute marketplace!\")'", "Command to run")
	flag.Parse()

	// Create Nakama client
	client := nakama.NewClient(*nakamaServer, *httpKey, *useSSL)

	// Authenticate with Nakama
	deviceID := uuid.New().String()
	session, err := client.AuthenticateDevice(context.Background(), deviceID, true, "buyer-sample")
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// Create a job request
	jobID := uuid.New().String()
	job := JobRequest{
		Image:   *image,
		Command: *command,
		BuyerID: session.UserID,
		JobID:   jobID,
	}

	// Convert job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("Failed to marshal job: %v", err)
	}

	// Send RPC to Nakama
	rpcResult, err := client.RpcFunc(context.Background(), session.Token, "send_job", string(jobData))
	if err != nil {
		log.Fatalf("Failed to send job: %v", err)
	}

	fmt.Printf("Job sent successfully! Job ID: %s\n", rpcResult.Payload)
	fmt.Println("This is a one-time job sender. Use buyerClient.go for receiving results.")
}
