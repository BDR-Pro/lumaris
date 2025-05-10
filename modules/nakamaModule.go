package modules

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	// Required for Nakama API client
	"github.com/heroiclabs/nakama-common/runtime"
	// Required for grpc.Dial
)

// InitModule initializes the Nakama module
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	// Register RPC function to handle job requests
	if err := initializer.RegisterRpc("send_job", SendJobToSeller); err != nil {
		logger.Error("Unable to register send_job RPC: %v", err)
		return err
	}

	// Register RPC function to handle job results
	if err := initializer.RegisterRpc("submit_job_result", SubmitJobResult); err != nil {
		logger.Error("Unable to register submit_job_result RPC: %v", err)
		return err
	}

	// Register RPC for sellers to register as available
	if err := initializer.RegisterRpc("register_seller", RegisterSeller); err != nil {
		logger.Error("Unable to register register_seller RPC: %v", err)
		return err
	}

	logger.Info("Compute marketplace module initialized")
	return nil
}

// SendJobToSeller forwards a job request to all sellers
func SendJobToSeller(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var job JobRequest
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		logger.Error("Failed to parse job request: %v", err)
		return "", errors.New("invalid job request format")
	}

	if job.Image == "" || job.Command == "" {
		return "", errors.New("job request must include image and command")
	}

	// Send job request to "sellers" topic
	topic := "sellers"

	// Create the job message
	jobMsg := map[string]interface{}{
		"type": "job_request",
		"data": job,
	}

	_, err := nk.ChannelMessageSend(ctx,
		topic,
		jobMsg,
		"job_request",
		"",
		false,
	)
	// Check if the message was sent successfully
	// In a real application, you would handle the error and retry logic
	if err != nil {
		logger.Error("Failed to send job to sellers: %v", err)
		return "", errors.New("failed to distribute job")
	}

	logger.Info("Job request sent to sellers. Job ID: %s", job.JobID)
	return job.JobID, nil
}

// SubmitJobResult processes the job result from a seller and notifies the buyer
func SubmitJobResult(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var result JobResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		logger.Error("Failed to parse job result: %v", err)
		return "", errors.New("invalid job result format")
	}

	if result.JobID == "" || result.BuyerID == "" || result.SellerID == "" {
		return "", errors.New("job result must include job_id, buyer_id, and seller_id")
	}

	// Send result to buyer via notification
	content := map[string]interface{}{
		"type": "job_result",
		"data": result,
	}

	notification := &runtime.NotificationSend{
		UserID:     result.BuyerID,
		Subject:    "Job Completed",
		Content:    content,
		Code:       1,
		Persistent: true,
	}

	if err := nk.NotificationsSend(ctx, []*runtime.NotificationSend{notification}); err != nil {
		logger.Error("Failed to send notification to buyer: %v", err)
		return "", errors.New("failed to notify buyer")
	}

	logger.Info("Job result sent to buyer. Job ID: %s", result.JobID)
	return "result_delivered", nil
}

// RegisterSeller is a dummy placeholder for seller registration
func RegisterSeller(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var seller struct {
		UserID       string   `json:"user_id"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(payload), &seller); err != nil {
		logger.Error("Failed to parse seller registration: %v", err)
		return "", errors.New("invalid seller registration format")
	}

	// In a real app, store this in a DB — this is a no-op placeholder
	logger.Info("Seller registered: %s with capabilities: %v", seller.UserID, seller.Capabilities)
	return "seller_registered", nil
}
