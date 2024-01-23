package batchr

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewInstance(t *testing.T) {
	batchProcessorFunc := func(items []int) {
		t.Logf("Received a batch of size %d", len(items))
		t.Logf("The batch is: %v", items)
	}

	capacityEvaluatorFunc := func(newItem int, existingItems []int) (r bool) {
		r = len(existingItems) == 5
		return
	}

	intervalEvaluatorFunc := func(lastUpdated *time.Time) (r bool) {
		if lastUpdated == nil {
			return false
		}
		cur := time.Now().UnixMilli()
		lu := lastUpdated.UnixMilli()

		return cur-lu > 700
	}

	b, err := New[int](
		batchProcessorFunc,
		capacityEvaluatorFunc,
		intervalEvaluatorFunc,
	)

	assrtNotNil(t, b)
	assrtNil(t, err)
	assrtTrue(t, b.IsEmpty())
	assrtEqual(t, 0, b.Size())

	b.Add(1)

	/*********************************
	 * The below tests may fail depending on
	 * processor speed/efficiency
	 * it's all about timing
	 **********************************/
	time.Sleep(50 * time.Millisecond)
	assrtFalse(t, b.IsEmpty()) // this may fail
	assrtEqual(t, 1, b.Size()) // this may fail

	b.Add(2, 3, 4, 5, 6, 7)
	assrtFalse(t, b.IsEmpty())

	time.Sleep(2 * time.Second) // allow time for the remaining of the batched to process (if any)
	b.Stop()
	assrtEqual(t, 2, b.BatchCount())

	//   **** Potential Test Output  ****
	// ../batchr/types_test.go:13: Received a batch of size 5
	// ../batchr/types_test.go:14: The batch is: [1 2 3 4 5]
	// ../batchr/types_test.go:13: Received a batch of size 2
	// ../batchr/types_test.go:14: The batch is: [6 7]
}

func TestAddAfterStop(t *testing.T) {
	batchProcessorFunc := func(items []int) {
		t.Logf("Received a batch of size %d", len(items))
		t.Logf("The batch is: %v", items)
	}

	capacityEvaluatorFunc := func(newItem int, existingItems []int) (r bool) {
		r = len(existingItems) == 5
		return
	}

	intervalEvaluatorFunc := func(lastUpdated *time.Time) (r bool) {
		if lastUpdated == nil {
			return false
		}
		cur := time.Now().UnixMilli()
		lu := lastUpdated.UnixMilli()

		return cur-lu > 700
	}

	b, err := New[int](
		batchProcessorFunc,
		capacityEvaluatorFunc,
		intervalEvaluatorFunc,
	)

	assrtNotNil(t, b)
	assrtNil(t, err)
	assrtTrue(t, b.IsEmpty())
	assrtEqual(t, 0, b.Size())

	assrtTrue(t, b.Add(1)) // SHOULD be added

	/*********************************
	 * The below tests may fail depending on
	 * processor speed/efficiency
	 * it's all about timing
	 **********************************/
	time.Sleep(50 * time.Millisecond)
	assrtFalse(t, b.IsEmpty()) // this may fail
	assrtEqual(t, 1, b.Size()) // this may fail

	assrtTrue(t,b.Add(2, 3, 4, 5, 6, 7)) // SHOULD be added
	assrtFalse(t, b.IsEmpty())

	time.Sleep(2 * time.Second) // allow time for the remaining of the batched to process (if any)
	b.Stop()
	assrtEqual(t, 2, b.BatchCount())

	assrtFalse(t, b.Add(999))	// should NOT be added

	assrtEqual(t, 2, b.BatchCount())

	//   **** Potential Test Output  ****
	// ../batchr/types_test.go:13: Received a batch of size 5
	// ../batchr/types_test.go:14: The batch is: [1 2 3 4 5]
	// ../batchr/types_test.go:13: Received a batch of size 2
	// ../batchr/types_test.go:14: The batch is: [6 7]
}

func Test_Stress01(t *testing.T) {
	var batchCount, itemCount, lastItemValue int

	batchCap := 1000
	expBatchCount := 1000
	expItemCount := expBatchCount * batchCap

	batchProcessorFunc := func(items []int) {
		t.Logf("Received a batch of size %d", len(items))
		batchCount++
		receivedItemsCount := len(items)
		itemCount += receivedItemsCount
		for _, iNum := range items {
			lastItemValue=iNum
		}
		if itemCount == expItemCount {
			t.Log("Just processed the last item!")
		}
	}

	capacityEvaluatorFunc := func(newItem int, existingItems []int) (r bool) {
		r = len(existingItems) == batchCap
		return
	}

	intervalEvaluatorFunc := func(lastUpdated *time.Time) (r bool) {
		if lastUpdated == nil {
			return false
		}
		cur := time.Now().UnixMilli()
		lu := lastUpdated.UnixMilli()

		timeSinceLastProcessed := cur-lu
		r = timeSinceLastProcessed > 700
		if r {
			t.Logf(
				"Need to force processing because too much time (%d ms) has passed since the last batch was processed",
				timeSinceLastProcessed)
		}
		return 
	}
	
	bOptions := Opts {
		PollingInterval: time.Second,
		NumChecksAfterStop: 3,
	}

	b, err := New(batchProcessorFunc, capacityEvaluatorFunc, intervalEvaluatorFunc, bOptions)
	assrtNil(t, err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var attemptedItemCount int
	for i := 0; i < expItemCount; i++ {
		if ! b.Add(i) {
			t.Logf("could not add item # %d to batcher",i)
		} else {
			attemptedItemCount++
		}
	}

	b.Stop()
	remainingItems := b.Size()
	t.Log("Stopped batcher")
	t.Logf("Attempted to add %d items", attemptedItemCount)
	t.Logf("Items remaining in batcher, after Stop(): %d",remainingItems)
	t.Logf("last item value to be processed: %d",lastItemValue)

	time.Sleep(1 * time.Second)

	t.Logf("Actual number of batches: %d", batchCount)
	assrtEqual(t, expBatchCount, batchCount)
	t.Logf("Actual number of items: %d", itemCount)
	actualItemCount := itemCount                                             //photo finish
	assrtEqualAny(t, []any{expItemCount, expItemCount - 1}, actualItemCount) // could be off due to polling interval and timing ...

	if actualItemCount == expItemCount-1 {
		time.Sleep(2 * time.Second) // delay until the final item is processed
		assrtTrue(t, b.IsEmpty(), "Batch should be empty, by now!")
	}

	assrtEqual(t, expItemCount, itemCount)
}


func Test_Stress02_RandomDelays(t *testing.T) {
	var batchCount, itemCount, lastItemValue int

	batchCap := 1000
	expBatchCount := 100
	expItemCount := expBatchCount * batchCap

	batchProcessorFunc := func(items []int) {
		t.Logf("Received a batch of size %d", len(items))
		batchCount++
		receivedItemsCount := len(items)
		itemCount += receivedItemsCount
		for _, iNum := range items {
			lastItemValue=iNum
		}
		if itemCount == expItemCount {
			t.Log("Just processed the last item!")
		}
	}

	capacityEvaluatorFunc := func(newItem int, existingItems []int) (r bool) {
		r = len(existingItems) == batchCap
		return
	}

	intervalEvaluatorFunc := func(lastUpdated *time.Time) (r bool) {
		if lastUpdated == nil {
			return false
		}
		cur := time.Now().UnixMilli()
		lu := lastUpdated.UnixMilli()

		timeSinceLastProcessed := cur-lu
		r = timeSinceLastProcessed > 700
		if r {
			t.Logf(
				"Need to force processing because too much time (%d ms) has passed since the last batch was processed",
				timeSinceLastProcessed)
		}
		return 
	}
	
	bOptions := Opts {
		PollingInterval: time.Second,
		NumChecksAfterStop: 3,
	}

	b, err := New(batchProcessorFunc, capacityEvaluatorFunc, intervalEvaluatorFunc, bOptions)
	assrtNil(t, err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var totalDelay time.Duration

	delay := func() {
		if totalDelay > 10 * time.Second {
			return
		}
		dVal := rand.Intn(900)
		dly := time.Duration(dVal) * time.Millisecond
		totalDelay += dly
		if dVal % 2 == 0 || totalDelay > 10 * time.Second {
			return
		}
		time.Sleep( dly)
	}

	var attemptedItemCount int
	for i := 0; i < expItemCount; i++ {
		if i % 1000 == 0 {
			delay()
		} 
		if ! b.Add(i) {
			t.Logf("could not add item # %d to batcher",i)
		} else {
			attemptedItemCount++
		}
	}

	b.Stop()
	remainingItems := b.Size()
	t.Log("Stopped batcher")
	t.Logf("Attempted to add %d items", attemptedItemCount)
	t.Logf("Items remaining in batcher, after Stop(): %d",remainingItems)
	t.Logf("last item value to be processed: %d",lastItemValue)

	time.Sleep(1 * time.Second)

	t.Logf("Actual number of batches: %d", batchCount)
	assrtEqual(t, expBatchCount, batchCount)
	t.Logf("Actual number of items: %d", itemCount)
	actualItemCount := itemCount                                             //photo finish
	assrtEqualAny(t, []any{expItemCount, expItemCount - 1}, actualItemCount) // could be off due to polling interval and timing ...

	if actualItemCount == expItemCount-1 {
		time.Sleep(2 * time.Second) // delay until the final item is processed
		assrtTrue(t, b.IsEmpty(), "Batch should be empty, by now!")
	}

	assrtEqual(t, expItemCount, itemCount)
}

func Test_Stress03_ConcurrentAdds(t *testing.T) {
	var batchCount, itemCount, lastItemValue int

	batchCap := 1000
	expBatchCount := 1000
	expItemCount := expBatchCount * batchCap

	batchProcessorFunc := func(items []int) {
		t.Logf("Received a batch of size %d", len(items))
		batchCount++
		receivedItemsCount := len(items)
		itemCount += receivedItemsCount
		for _, iNum := range items {
			lastItemValue=iNum
		}
		if itemCount == expItemCount {
			t.Log("Just processed the last item!")
		}
	}

	capacityEvaluatorFunc := func(newItem int, existingItems []int) (r bool) {
		r = len(existingItems) == batchCap
		return
	}

	intervalEvaluatorFunc := func(lastUpdated *time.Time) (r bool) {
		if lastUpdated == nil {
			return false
		}
		cur := time.Now().UnixMilli()
		lu := lastUpdated.UnixMilli()

		timeSinceLastProcessed := cur-lu
		r = timeSinceLastProcessed > 700
		if r {
			t.Logf(
				"Need to force processing because too much time (%d ms) has passed since the last batch was processed",
				timeSinceLastProcessed)
		}
		return 
	}
	
	bOptions := Opts {
		PollingInterval: time.Second,
		NumChecksAfterStop: 3,
	}

	b, err := New(batchProcessorFunc, capacityEvaluatorFunc, intervalEvaluatorFunc, bOptions)
	assrtNil(t, err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var attemptedItemCount int

	var w sync.WaitGroup
	w.Add(2)
	go func (iCount *int)  {
		defer w.Done()
		for j := 0; j < expItemCount; j+=2 {
			if ! b.Add(j*-1) {
				t.Logf("(p) could not add item # %d to batcher",j)
			} else {
				*iCount++
			}
		}
	}(&attemptedItemCount)
	for i := 1; i < expItemCount; i+=2 {
		if ! b.Add(i) {
			t.Logf("could not add item # %d to batcher",i)
		} else {
			attemptedItemCount++
		}
	}
	w.Done()

	// wait for concurrent processing to complete before stopping
	w.Wait()

	b.Stop()
	remainingItems := b.Size()
	t.Log("Stopped batcher")
	t.Logf("Attempted to add %d items", attemptedItemCount)
	t.Logf("Items remaining in batcher, after Stop(): %d",remainingItems)
	t.Logf("last item value to be processed: %d",lastItemValue)

	time.Sleep(1 * time.Second)
	t.Logf("Actual number of batches: %d", batchCount)
	assrtEqual(t, expBatchCount, batchCount)
	t.Logf("Actual number of items: %d", itemCount)
	actualItemCount := itemCount                                             //photo finish
	assrtEqualAny(t, []any{expItemCount, expItemCount - 1}, actualItemCount) // could be off due to polling interval and timing ...

	assrtEqual(t, expItemCount, itemCount)
}
