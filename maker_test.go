package batchr

import (
	"math/rand"
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

	delay := func() {
		dVal := rand.Intn(900)
		if dVal % 2 == 0 {
			return
		}
		time.Sleep( time.Duration(dVal) * time.Millisecond)
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

const tFail = "test failed"

func failTest(t *testing.T, msgAndArgs0 ...any) {
	t.Helper()

	if len(msgAndArgs0) == 0 {
		t.Errorf(tFail)
	}
	var msg string
	var msgOK bool
	msg, msgOK = msgAndArgs0[0].(string)
	if !msgOK {
		t.Errorf(tFail)
	}
	args := msgAndArgs0[1:]
	t.Errorf(msg, args...)
}

func assrtEqual(t *testing.T, oe, oa any, msgAndArgs0 ...any) {
	t.Helper()

	if oe == oa {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected values to be equal, but\n\texpected value was %v\n\tactual value was %v", oe, oa)
		return
	}
	failTest(t, msgAndArgs0)
}

func assrtEqualAny(t *testing.T, oe0 []any, oa any, msgAndArgs0 ...any) {
	t.Helper()
	if len(oe0) == 0 {
		t.Errorf("impossible - empty assertion for EqualAny")
		return
	}
	for _, oe := range oe0 {
		if oe == oa {
			return
		}
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected value to equal one of %v, but\n\tactual value was %v", oe0, oa)
		return
	}
	failTest(t, msgAndArgs0)
}

func assrtFalse(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == false {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected bool false, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0)
}

func assrtTrue(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == true {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected bool true, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0)
}

func assrtNil(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == nil {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected nil, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0)
}

func assrtNotNil(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o != nil {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected non-nil, but actual value was nil")
		return
	}
	failTest(t, msgAndArgs0)
}
