// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package weight

import (
	"context"
)

const (
	deadQueueSize = 512
	maxQueueSize  = 10000
)

//go:generate mockgen -source=dlq.go -destination=dlq_mock.go -package=weight
type (
	// DLQ (Dead Letter Queue) interface provides functionality to handle failed events.
	// It allows pushing new failed events into the queue, popping events for retry,
	// checking the size of the queue, retrieving all events, removing specific events,
	// and closing the queue when it's no longer needed.
	DLQ interface {
		// Push adds a new event to the dead letter queue.
		// ctx: Context for handling cancellation or timeouts.
		// event: The DLQEvent to be stored in the queue.
		// Returns an error if the operation fails.
		Push(ctx context.Context, event DLQEvent) error

		// Pop retrieves and removes the oldest event from the queue.
		// ctx: Context for handling cancellation or timeouts.
		// Returns the retrieved DLQEvent and an error if the operation fails.
		Pop(ctx context.Context) (event DLQEvent, err error)

		// GetSize returns the current number of events in the queue.
		// Returns the size as an int and an error if the operation fails.
		GetSize() (int, error)

		// GetAll retrieves all events currently in the queue without removing them.
		// ctx: Context for handling cancellation or timeouts.
		// Returns a slice of DLQEvents and an error if the operation fails.
		GetAll(ctx context.Context) ([]DLQEvent, error)

		// Remove deletes a specific event from the queue by its ID.
		// ctx: Context for handling cancellation or timeouts.
		// id: The unique identifier of the event to be removed.
		// Returns an error if the operation fails.
		Remove(ctx context.Context, id int64) error

		// Close performs cleanup operations and safely shuts down the queue.
		Close()
	}

	// PersistentDLQ extends the DLQ interface with additional functionality for recovering events
	// from persistent storage. This interface is useful in scenarios where queue state needs to be
	// preserved across system restarts or failures.
	PersistentDLQ interface {
		DLQ

		// Recover retrieves all persisted events from the storage that were not successfully processed.
		// Returns a slice of DLQEvents that need to be retried and an error if the recovery operation fails.
		Recover() ([]DLQEvent, error)
	}

	// DistributedDLQ extends the PersistentDLQ interface to provide functionality for distributed environments.
	// This interface ensures that dead letter queue operations can be coordinated across multiple nodes in a cluster.
	DistributedDLQ interface {
		PersistentDLQ

		// ClusterSize returns the number of nodes currently participating in the distributed queue system.
		// This information can be used for load balancing or determining replication factors.
		ClusterSize() int
	}

	// DLQEvent represents a failed event that is stored in the Dead Letter Queue (DLQ).
	// It contains detailed information about the failure, including the original event,
	// reason for failure, error details, retry attempts, and timestamps.
	DLQEvent struct {
		// Unique identifier for the event
		ID int64
		// The original event that failed processing
		OriginalEvent Event
		// Human-readable description of why the event failed
		FailReason string
		// The actual error encountered during processing
		Err error
		// Number of times this event has been retried
		RetryTimes int
		// Time when the event was first added to the DLQ
		Timestamp int64
		// Time of the most recent retry attempt
		LastAttempt int64
	}
)
