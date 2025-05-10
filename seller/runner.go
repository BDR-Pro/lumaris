package seller

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bdr-pro/lumaris/modules"
	"github.com/google/uuid"
)

// RunnerMain is the entry point for the seller runner
func RunnerMain() {
	nakamaServer := flag.String("server", "127.0.0.1:7350", "Nakama server address")
	sessionToken := flag.String("token", "", "Nakama session token")
	flag.Parse()

	if *sessionToken == "" {
		log.Fatal("You must provide a session token using -token")
	}

	// Check Docker availability
	if err := checkDocker(); err != nil {
		log.Fatalf("Docker not available: %v", err)
	}

	sellerID := "seller-id-placeholder" // Optionally decode this from token or keep static

	registerSeller(*nakamaServer, *sessionToken, sellerID)

	// Example: simulate receiving a job from queue or polling
	go func() {
		time.Sleep(2 * time.Second)
		job := modules.JobRequest{
			Image:   "python:3.10",
			Command: "python -c 'print(\"Hello from Seller\")'",
			BuyerID: "buyer123",
			JobID:   uuid.New().String(),
		}
		executeJob(*nakamaServer, *sessionToken, sellerID, job)
	}()

	// Wait for CTRL+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Seller shutting down.")
}

func checkDocker() error {
	cmd := exec.Command("docker", "--version")
	return cmd.Run()
}

func registerSeller(server, token, sellerID string) {
	payload := map[string]interface{}{
		"user_id":      sellerID,
		"capabilities": []string{"python:3.10", "node:16", "ubuntu:latest"},
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v2/rpc/register_seller", server), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error building register request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to register seller: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Registration failed [%d]: %s", resp.StatusCode, body)
	}

	log.Println("Seller registered successfully.")
}

func executeJob(server, token, sellerID string, job modules.JobRequest) {
	log.Printf("Executing job: %s using image: %s", job.JobID, job.Image)

	result := modules.JobResult{
		JobID:     job.JobID,
		BuyerID:   job.BuyerID,
		SellerID:  sellerID,
		Timestamp: time.Now().Unix(),
	}

	cmd := exec.Command("docker", "run", "--rm", "--network=none", "--memory=512m", "--cpus=1", job.Image, "sh", "-c", job.Command)
	output, err := cmd.CombinedOutput()

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

	resultJSON, _ := json.Marshal(result)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v2/rpc/submit_job_result", server), bytes.NewBuffer(resultJSON))
	if err != nil {
		log.Printf("Failed to build result request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to submit job result: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Job result rejected [%d]: %s", resp.StatusCode, body)
		return
	}

	log.Printf("Job %s submitted successfully", job.JobID)
}
