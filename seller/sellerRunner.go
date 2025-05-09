package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/heroiclabs/nakama-common/rtapi"
	nakama "github.com/heroiclabs/nakama-go"
)

// JobRequest defines the structure of a compute job
type JobRequest struct {
	Image   string `json:"image"`    // Docker image to use
	Command string `json:"command"`  // Command to run inside the container
	BuyerID string `json:"buyer_id"` // ID of the buyer requesting the job
	JobID   string `json:"job_id"`   // Unique identifier for the job
}

// JobResult contains the result of a job execution
type JobResult struct {
	JobID     string `json:"job_id"`    // ID of the job that was executed
	BuyerID   string `json:"buyer_id"`  // ID of the buyer who requested the job
	SellerID  string `json:"seller_id"` // ID of the seller who executed the job
	Output    string `json:"output"`    // Output from the command execution
	Error     string `json:"error"`     // Error message if job failed
	ExitCode  int    `json:"exit_code"` // Exit code from the container
	Timestamp int64  `json:"timestamp"` // When the job was completed
}

func main() {
	// Parse command line flags
	nakamaServer := flag.String("server", "127.0.0.1:7350", "Nakama server address")
	httpKey := flag.String("key", "defaultkey", "Nakama server HTTP key")
	useSSL := flag.Bool("ssl", false, "Use SSL for connection")
	flag.Parse()

	// Check if Docker is available
	if err := checkDocker(); err != nil {
		log.Fatalf("Docker is not available: %v", err)
	}

	// Create Nakama client
	client := nakama.NewClient(*nakamaServer, *httpKey, *useSSL)

	// Authenticate with Nakama
	deviceID := uuid.New().String()
	session, err := client.AuthenticateDevice(context.Background(), deviceID, true, "seller")
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	log.Printf("Authenticated as seller: %s", session.UserID)

	// Create a realtime client
	rtClient, err := client.NewRtClient(session)
	if err != nil {
		log.Fatalf("Failed to create realtime client: %v", err)
	}

	// Create context with cancel for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the realtime client
	rtClient.Connect(ctx, 5)

	// Join the sellers topic
	topic := "sellers"
	if _, err := rtClient.ChannelJoin(topic, false); err != nil {
		log.Fatalf("Failed to join sellers topic: %v", err)
	}
	log.Printf("Joined sellers topic")

	// Register for job requests
	rtClient.OnChannelMessage(func(channelMessage *rtapi.ChannelMessage) {
		// Parse the message
		var message struct {
			Type string     `json:"type"`
			Data JobRequest `json:"data"`
		}

		if err := json.Unmarshal([]byte(channelMessage.Content), &message); err != nil {
			log.Printf("Failed to parse channel message: %v", err)
			return
		}

		// Handle job request
		if message.Type == "job_request" {
			job := message.Data
			log.Printf("Received job request: %s", job.JobID)

			// Execute the job asynchronously
			go func() {
				executeJob(ctx, client, session.Token, session.UserID, job)
			}()
		}
	})

	// Register seller capabilities with Nakama
	registerSeller(ctx, client, session.Token, session.UserID)

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

// checkDocker verifies that Docker is available
func checkDocker() error {
	cmd := exec.Command("docker", "--version")
	return cmd.Run()
}

// registerSeller registers the seller's capabilities with Nakama
func registerSeller(ctx context.Context, client *nakama.Client, token, sellerID string) {
	// Get supported Docker images
	// In a real implementation, you might have a more sophisticated way to determine capabilities
	capabilities := []string{"python:3.10", "node:16", "ubuntu:latest"}

	// Create registration payload
	registration := struct {
		UserID       string   `json:"user_id"`
		Capabilities []string `json:"capabilities"`
	}{
		UserID:       sellerID,
		Capabilities: capabilities,
	}

	// Convert to JSON
	data, err := json.Marshal(registration)
	if err != nil {
		log.Fatalf("Failed to marshal registration: %v", err)
	}

	// Register with Nakama
	_, err = client.RpcFunc(ctx, token, "register_seller", string(data))
	if err != nil {
		log.Fatalf("Failed to register seller: %v", err)
	}

	log.Printf("Registered as seller with capabilities: %v", capabilities)
}

// executeJob runs a Docker container and returns the result
func executeJob(ctx context.Context, client *nakama.Client, token, sellerID string, job JobRequest) {
	log.Printf("Executing job %s with image %s", job.JobID, job.Image)

	// Create the result structure
	result := JobResult{
		JobID:     job.JobID,
		BuyerID:   job.BuyerID,
		SellerID:  sellerID,
		Timestamp: time.Now().Unix(),
	}

	// Execute Docker command
	cmd := exec.Command("docker", "run", "--rm", "--network=none", "--memory=512m", "--cpus=1", job.Image, "sh", "-c", job.Command)
	output, err := cmd.CombinedOutput()

	// Process the result
	result.Output = string(output)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err.Error()
	} else {
		result.ExitCode = 0
	}

	// Convert result to JSON
	resultData, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}

	// Send result back to Nakama
	_, err = client.RpcFunc(ctx, token, "submit_job_result", string(resultData))
	if err != nil {
		log.Printf("Failed to submit job result: %v", err)
		return
	}

	log.Printf("Job %s completed with exit code %d", job.JobID, result.ExitCode)
}
