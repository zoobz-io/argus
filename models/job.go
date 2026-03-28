package models

import "time"

// JobStatus represents the state of an ingestion job.
type JobStatus string

// Job status values.
const (
	JobPending    JobStatus = "pending"
	JobProcessing JobStatus = "processing"
	JobCompleted  JobStatus = "completed"
	JobFailed     JobStatus = "failed"
)

// Job tracks the ingestion processing of a document version.
type Job struct {
	CreatedAt  time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Error      *string   `json:"error,omitempty" db:"error"`
	Status     JobStatus `json:"status" db:"status" constraints:"notnull" default:"'pending'"`
	ID         string    `json:"id" db:"id" constraints:"primarykey"`
	VersionID  string    `json:"version_id" db:"version_id" constraints:"notnull"`
	DocumentID string    `json:"document_id" db:"document_id" constraints:"notnull"`
	TenantID   string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
}

// GetID returns the job's primary key.
func (j Job) GetID() string {
	return j.ID
}

// Clone returns a deep copy of the job.
func (j Job) Clone() Job {
	c := j
	if j.Error != nil {
		e := *j.Error
		c.Error = &e
	}
	return c
}
