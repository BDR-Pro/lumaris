package modules

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
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

// SendJobToSeller forwards a job request to an available seller
func SendJobToSeller(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, payload string) (string, error) {
	// Parse the job request
	var job JobRequest
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		logger.Error("Failed to parse job request: %v", err)
		return "", errors.New("invalid job request format")
	}

	// Validate job request
	if job.Image == "" || job.Command == "" {
		return "", errors.New("job request must include image and command")
	}

	// Get available sellers (in a real implementation, you would have a more sophisticated selection process)
	// For now, we'll broadcast to all sellers and let them decide if they want to take the job
	// In a real system, you'd track seller availability and capabilities in a database

	// Create a topic for sellers
	topic := "sellers"

	// Create a JSON message for sellers
	jobMsg, err := json.Marshal(map[string]interface{}{
		"type": "job_request",
		"data": job,
	})
	if err != nil {
		logger.Error("Failed to marshal job message: %v", err)
		return "", errors.New("internal server error")
	}

	// Send the job to all sellers in the topic
	if err := nk.ChannelMessageSend(ctx, topic, string(jobMsg)); err != nil {
		logger.Error("Failed to send job to sellers: %v", err)
		return "", errors.New("failed to distribute job")
	}

	logger.Info("Job request sent to sellers. Job ID: %s", job.JobID)

	// Return job ID for tracking
	return job.JobID, nil
}

// SubmitJobResult processes a job result from a seller
func SubmitJobResult(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, payload string) (string, error) {
	// Parse the job result
	var result JobResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		logger.Error("Failed to parse job result: %v", err)
		return "", errors.New("invalid job result format")
	}

	// Validate result
	if result.JobID == "" || result.BuyerID == "" || result.SellerID == "" {
		return "", errors.New("job result must include job_id, buyer_id, and seller_id")
	}

	// Store the result (in a production system, you'd likely store this in a database)
	// Here we're just forwarding it to the buyer

	// Send direct message to buyer with the result
	content, err := json.Marshal(map[string]interface{}{
		"type": "job_result",
		"data": result,
	})
	if err != nil {
		logger.Error("Failed to marshal result message: %v", err)
		return "", errors.New("internal server error")
	}

	// Send notification to buyer
	notifications := []*runtime.NotificationSend{
		{
			UserID:     result.BuyerID,
			Subject:    "Job Completed",
			Content:    string(content),
			Code:       1, // Custom code for job results
			Persistent: true,
		},
	}

	if err := nk.NotificationsSend(ctx, notifications); err != nil {
		logger.Error("Failed to send notification to buyer: %v", err)
		return "", errors.New("failed to notify buyer")
	}

	logger.Info("Job result sent to buyer. Job ID: %s", result.JobID)

	return "result_delivered", nil
}

// RegisterSeller registers a seller as available for jobs
func RegisterSeller(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, payload string) (string, error) {
	// Parse seller registration
	var seller struct {
		UserID       string   `json:"user_id"`
		Capabilities []string `json:"capabilities"` // e.g., supported Docker images
	}

	if err := json.Unmarshal([]byte(payload), &seller); err != nil {
		logger.Error("Failed to parse seller registration: %v", err)
		return "", errors.New("invalid seller registration format")
	}

	// In a real implementation, you would store seller information in a database
	// For now, we'll just join them to the sellers topic

	// Join the seller to the sellers topic
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		logger.Error("Invalid context user ID")
		return "", errors.New("authentication required")
	}

	topic := "sellers"
	if _, err := nk.ChannelJoin(ctx, topic, userID, false); err != nil {
		logger.Error("Failed to join seller to topic: %v", err)
		return "", errors.New("failed to register seller")
	}

	logger.Info("Seller registered: %s", userID)

	return "seller_registered", nil
}
