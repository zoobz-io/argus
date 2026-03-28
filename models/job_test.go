package models

import "testing"

func TestJob_GetID(t *testing.T) {
	j := Job{ID: "job-1"}
	if got := j.GetID(); got != "job-1" {
		t.Errorf("GetID() = %q, want %q", got, "job-1")
	}
}

func TestJob_Clone(t *testing.T) {
	errMsg := "something failed"
	j := Job{
		ID:     "job-1",
		Status: models_JobFailed(),
		Error:  &errMsg,
	}

	clone := j.Clone()

	if clone.ID != j.ID {
		t.Errorf("ID: got %q, want %q", clone.ID, j.ID)
	}

	// Pointer independence.
	if clone.Error == j.Error {
		t.Error("Error: clone shares pointer with original")
	}
	if *clone.Error != errMsg {
		t.Errorf("Error value: got %q, want %q", *clone.Error, errMsg)
	}

	// Mutate clone, verify original unchanged.
	*clone.Error = "mutated"
	if *j.Error != errMsg {
		t.Error("mutating clone leaked to original")
	}
}

func TestJob_Clone_NilError(t *testing.T) {
	j := Job{ID: "job-1"}
	clone := j.Clone()
	if clone.Error != nil {
		t.Error("expected nil Error in clone")
	}
}

// helper to avoid using the constant directly in struct literal (lint).
func models_JobFailed() JobStatus { return JobFailed }
