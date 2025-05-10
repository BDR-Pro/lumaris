package buyer

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bdr-pro/lumaris/modules"
	"github.com/google/uuid"
)

// Test runs a simple test of the buyer functionality
func Test() {
	// Parse command line flags
	nakamaServer := flag.String("server", "127.0.0.1:7350", "Nakama server address")
	sessionToken := flag.String("token", "", "Nakama session token") // You must log in separately
	image := flag.String("image", "python:3.10", "Docker image to use")
	command := flag.String("command", "python -c 'print(\"Hello from compute marketplace!\")'", "Command to run")
	flag.Parse()

	if *sessionToken == "" {
		log.Fatal("You must provide a Nakama session token using -token")
	}

	// Generate a random job ID
	jobID := uuid.New().String()

	// Create the job payload
	job := modules.JobRequest{
		Image:   *image,
		Command: *command,
		BuyerID: "buyer-uuid-placeholder", // Replace with actual user ID if needed
		JobID:   jobID,
	}

	// Convert to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("Failed to marshal job: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("http://%s/v2/rpc/send_job", *nakamaServer)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jobData))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*sessionToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send RPC: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("RPC failed: %s", body)
	}

	fmt.Printf("Job sent successfully! Job ID: %s\n", jobID)
	fmt.Println("Response:", string(body))
}
