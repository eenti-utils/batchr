// Package batchr facilitates processing data in batches
package batchr

import "time"

// Accepts items of the specified type and processes the items in batches of the specified size,
// using the specified batch processor function
type Batcher[V any] interface {
	// returns bool true, if the batch holder does not have any items
	IsEmpty() bool
	// accepts items to be processed by the specified batch processor function, if the batcher is not halted
	Add(items0 ...V) bool
	// cease all polling, acceptance, and processing of new items. process remaining items in the batch holder, if any.
	Stop()
	// returns the number of items currently in the batch holder
	Size() int
	// returns the number of batches that have been processed
	BatchCount() int
}

// a function that processes a batch of items
type BatchProcessor[V any] func(items []V)

// a function that determines if the batch is at capacity, based on the
// existing items in the batch, and the newly submitted item
//   - returns bool true, if the batch is at capacity, and bool false, otherwise
//   - when true, existing items are passed to the BatchProcessor function,
//     and the newly submitted item is added to the new batch
type CapacityEvaluator[V any] func(newItem V, existingItems []V) bool

// a function that determines if the current batch should be sent to the
// BatchProcessor function, based on the amount of time passed, since its most
// recent update.
//   - returns bool true, if existing items from the current batch should be
//     passed to the BatchProcessor function (regardless of capacity,) false,
//     otherwise
//   - this function prevents items from becoming "stuck", which might otherwise
//     happen, when a batch does not reach capacity, and submitters have stopped
//     sending new items to the Batcher
type IntervalEvaluator func(lastUpdated *time.Time) bool
