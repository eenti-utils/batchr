# batchr

The batchr package is a library written in Go that facilitates the processing of data (any type) in batches.

This is accomplished by:
1. instantiating a `Batcher` object of the user-specified data type
1. feeding elements of the user-specified data type to the `Batcher` object

**Usage Example:**
```go
  var myBatcher batchr.Batcher[Cupcake]
  // ... assuming we've instantiated the myBatcher object to process the user-defined type Cupcake, in batches of 10 

  var cupcakeArr []Cupcake
  // ... assuming that cupcakeArr now has several zillion Cupcake elements

  // Example 1
  myBatcher.Add(cupcakeArr...)

  // Example 2
  for _, c := range cupcakeArr {
    myBatcher.Add(c)
  }

  /**********************************************************************************************
   * How does myBatcher know to make batches of 10 (and not 11, or some other arbitrary value)? *
   * What does myBatcher do with the batch, once it has the correct number of cupcakes??        *
   **********************************************************************************************/
```

Instantiating a `batchr.Batcher` instance
----

In the above usage example, we skipped over instantiation

To instantiate a `batchr.Batcher` instance, the following functions are needed
- processor - this function defines what is done with the batch of items, once it is obtained
- capacity evaluator - this function simply returns bool `true` when the batch has reached the desired capacity
- time interval evaluator - this function is a failsafe that returns bool `true` if "too much" time has passed since the last batch was processed

  

**Processor Function**

**type:** `batchr.BatchProcessor[V any] func(items []V)`

```go
  // What the Cupcake Batcher does with the batch

  // assuming we have a batch of cupcakes, ready to go ...
  func packageAndDeliver(cakes []Cupcake) {
    // ... do great things, here ...
  }
```

**Capacity Evaluator Function**

**type:** `batchr.CapacityEvaluator[V any] func(newItem V, existingItems []V) bool`

```go
  // How the Cupcake Batcher knows to make batches of 10 (and not 11, or some other arbitrary value)

  // is the batch "full", yet? ...
  func checkTheBatchSize(newCake Cupcake, existingCakesInCurrentBatch []Cupcake) bool {
    // the "newCake" variable is not needed, in this example ...

    return len(existingCakesInCurrentBatch) == 10
  }
```

**Time Interval Evaluator Function**

**type:** `batchr.IntervalEvaluator func(lastUpdated *time.Time) bool`

```go
  // so that cupcakes don't get left or "stuck" in the batcher ...
  func isItTimeForANewBatch(lastUpdated *time.Time) bool {
    if lastUpdated == nil {
      return false
    }

    // arbitrarily, we'll say that 3000 ms is the max we'd want to wait before processing a batch of cupcakes
    var tooMuchTime int64 = 3000  
    cur := time.Now().UnixMilli()
    lu := lastUpdated.UnixMilli()
    return (cur - lu) > tooMuchTime
  }
```

**Instantiation Example:**

**type:** `batchr.Batcher[V any] interface`

```go
  /****************************************************************************************************
  *  assuming the previously defined functions:                                                       *
  *   func packageAndDeliver(cakes []Cupcake) ...                                                     *
  *   func checkTheBatchSize(newCake Cupcake, existingCakesInCurrentBatch []Cupcake) bool ...         *
  *   func isItTimeForANewBatch(lastUpdated *time.Time) bool ...                                      *
  *****************************************************************************************************/
  myBatcher, err := batchr.New[Cupcake](packageAndDeliver, checkTheBatchSize, isItTimeForANewBatch)

  if err != nil {
    panic(err)
  }

  // now add items to be processed in batches ...
  myBatcher.Add(...)

  // concurrent adds are also supported ...
  go func() {
    myBatcher.Add(...)  // this is OK to do
  }()
```
