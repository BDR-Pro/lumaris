package modules

// JobRequest represents a compute job to be executed
type JobRequest struct {
	Image   string `json:"image"`    // Docker image to use
	Command string `json:"command"`  // Command to run inside the container
	BuyerID string `json:"buyer_id"` // ID of the buyer requesting the job
	JobID   string `json:"job_id"`   // Unique identifier for the job
}

// JobResult represents the result of a compute job
type JobResult struct {
	JobID     string `json:"job_id"`    // ID of the job that was executed
	BuyerID   string `json:"buyer_id"`  // ID of the buyer who requested the job
	SellerID  string `json:"seller_id"` // ID of the seller who executed the job
	Output    string `json:"output"`    // Output from the command execution
	Error     string `json:"error"`     // Error message if job failed
	ExitCode  int    `json:"exit_code"` // Exit code from the container
	Timestamp int64  `json:"timestamp"` // When the job was completed
}
