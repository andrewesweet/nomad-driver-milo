//go:build e2e

package e2e

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
)

// JobRunner manages job submission and monitoring
type JobRunner struct {
	client *api.Client
}

// NewJobRunner creates a new JobRunner with the given Nomad API client
func NewJobRunner(client *api.Client) *JobRunner {
	return &JobRunner{
		client: client,
	}
}

// ValidateJobSpec validates that a job specification has required fields
func (jr *JobRunner) ValidateJobSpec(job *api.Job) error {
	if job == nil {
		return fmt.Errorf("job specification is nil")
	}
	
	if job.ID == nil || *job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	
	if job.Name == nil || *job.Name == "" {
		return fmt.Errorf("job name is required")
	}
	
	if job.Type == nil || *job.Type == "" {
		return fmt.Errorf("job type is required")
	}
	
	return nil
}