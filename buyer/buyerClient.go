package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/heroiclabs/nakama-common/api"
	nakama "github.com/heroiclabs/nakama-go"
)

// Job structure matching the server-side JobRequest
type JobRequest struct {
	Image   string `json:"image"`    // Docker image to use
	Command string `json:"command"`  // Command to run inside the container
	BuyerID string `json:"buyer_id"` // ID of the buyer requesting the job
	JobID   string `json:"job_id"`   // Unique identifier for the job
}

// JobResult structure matching the server-side JobResult
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

	// Create Nakama client
	client := nakama.NewClient(*nakamaServer, *httpKey, *useSSL)

	// Create a device ID
	deviceID := uuid.New().String()

	// Authenticate with Nakama
	session, err := client.AuthenticateDevice(context.Background(), deviceID, true, "buyer")
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	log.Printf("Authenticated as: %s", session.UserID)

	// Create a realtime client to receive notifications
	rtClient, err := client.NewRtClient(session)
	if err != nil {
		log.Fatalf("Failed to create realtime client: %v", err)
	}

	// Create context with cancel for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the realtime client
	rtClient.Connect(ctx, 5)

	// Register notification handler
	rtClient.OnNotification(func(notification *api.Notification) {
		log.Printf("Received notification: %s", notification.Subject)
		if notification.Code == 1 { // Code 1 is for job results
			var message struct {
				Type string    `json:"type"`
				Data JobResult `json:"data"`
			}

			if err := json.Unmarshal([]byte(notification.Content), &message); err != nil {
				log.Printf("Failed to parse notification: %v", err)
				return
			}

			if message.Type == "job_result" {
				result := message.Data
				log.Printf("Job completed: %s", result.JobID)
				log.Printf("Output: %s", result.Output)
				if result.Error != "" {
					log.Printf("Error: %s", result.Error)
				}
				log.Printf("Exit code: %d", result.ExitCode)
			}
		}
	})

	// Load and send a sample job
	sendSampleJob(ctx, client, session.Token, session.UserID)

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

// sendSampleJob sends a sample job request to Nakama
func sendSampleJob(ctx context.Context, client *nakama.Client, token, userID string) {
	// Create a job request
	jobID := uuid.New().String()
	job := JobRequest{
		Image:   "python:3.10",
		Command: "python -c 'print(\"Hello from compute marketplace!\")'",
		BuyerID: userID,
		JobID:   jobID,
	}

	// Convert job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("Failed to marshal job: %v", err)
	}

	// Send RPC to Nakama
	rpcResult, err := client.RpcFunc(ctx, token, "send_job", string(jobData))
	if err != nil {
		log.Fatalf("Failed to send job: %v", err)
	}

	log.Printf("Job sent successfully! Job ID: %s", rpcResult.Payload)
	log.Printf("Waiting for job result...")
}
