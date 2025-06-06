package pool

import (
	"github.com/google/uuid"
)

type HandleError struct {
	workerID    uuid.UUID
	jobID       uuid.UUID
	description string
	err         error
}

func NewHandleError(workerID uuid.UUID, jobID uuid.UUID, description string, err error) HandleError {
	return HandleError{
		workerID:    workerID,
		jobID:       jobID,
		description: description,
		err:         err,
	}
}

func (h HandleError) Error() string {
	return h.err.Error()
}
