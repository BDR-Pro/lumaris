package buyer

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bdr-pro/lumaris/modules"
	"github.com/google/uuid"
)

// ClientMain is the REST-based client entry point
func ClientMain() {
	// Parse flags
	nakamaServer := flag.String("server", "127.0.0.1:7350", "Nakama server address")
	sessionToken := flag.String("token", "", "Nakama session token")
	flag.Parse()

	if *sessionToken == "" {
		log.Fatal("You must provide a session token using -token")
	}

	userID := "buyer-user-id-placeholder" // Replace if needed

	// Create job
	jobID := uuid.New().String()
	job := modules.JobRequest{
		Image:   "python:3.10",
		Command: "python -c 'print(\"Hello from compute marketplace!\")'",
		BuyerID: userID,
		JobID:   jobID,
	}

	log.Printf("Sending job with ID: %s\n", jobID)
	err := sendJobRequest(*nakamaServer, *sessionToken, job)
	if err != nil {
		log.Fatalf("Failed to send job: %v", err)
	}
	log.Println("Job sent successfully. Waiting for result...")

	// Simulate wait for result (e.g., polling, long-polling, or notification system)
	waitForResult(jobID)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down.")
}

// sendJobRequest sends the job as an RPC request using Nakama's REST API
func sendJobRequest(server, token string, job modules.JobRequest) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	url := fmt.Sprintf("http://%s/v2/rpc/send_job", server)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jobData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send RPC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RPC failed [%d]: %s", resp.StatusCode, string(body))
	}

	return nil
}

// waitForResult is a placeholder for waiting on job results.
// You can later connect this to a websocket or polling endpoint.
func waitForResult(jobID string) {
	log.Printf("Pretending to wait for job result... [Job ID: %s]", jobID)
	time.Sleep(3 * time.Second) // Simulate waiting
	log.Printf("Job %s finished. (Result fetching not implemented)", jobID)
}
