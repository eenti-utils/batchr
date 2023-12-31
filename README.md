# batchr

The batchr package is a library written in Go that facilitates the processing of data (any type) in batches.

This is accomplished by:
1. instantiating a `batchr.Batcher` object of the user-specified data type
1. feeding items of the user-specified data type to the `batchr.Batcher` object

**Ideal Use Case:**

A `batchr.Batcher` object is ideal for use in long-running processes where large numbers of items of a given type must be processed in groups, at a time.

- several zillion items to be processed in groups of umteen items, each
- thousands file objects to be processed in groups where the total number of bytes for all files in the group is less than 500MB

For the duration of the long-running process, items are fed to the `batchr.Batcher` object via its Add(...) method: 
- at random
- at once (i.e. as an array or slice)
- in sequence
- concurrently
- all (or some combination) of the above

The `batchr.Batcher` object collects the items into groups (i.e. _batches_,) and processes each group.

The examples following use a user-defined data type called `Cupcake` as the item type.

**Usage Example:**
```go
  var myBatcher batchr.Batcher[Cupcake]
  /*************************************
   * This example assumes that  we've  *
   * instantiated the myBatcher object *
   * to process the user-defined type  *
   * Cupcake, in batches of 10         * 
   *************************************/

  var cupcakeArr []Cupcake
  // ... assuming that cupcakeArr now has several zillion Cupcake elements

  // Example 1
  myBatcher.Add(cupcakeArr...)

  // Example 2
  for _, c := range cupcakeArr {
    myBatcher.Add(c)
  }

  /**************************************
   * 1. How does myBatcher know to make *
   * batches of 10 (and not 11, or some *
   * other arbitrary value)?            *
   *                                    *
   * 2. What does myBatcher do with the *
   * batch, once it has the correct     *
   * number of cupcakes??               *
   **************************************/
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
    // the "newCake" parameter is not needed, in this example ...

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

**Options:**

**type:** `batchr.Opts struct`

By default, the `batchr.Batcher` object: 
- has a polling interval of 1 second
- will check three (3) times, after being stopped (so that items are not left in the batcher) 

Any of these values may be adjusted by passing in the `batchr.Opts` struct 
as the final parameter to the `batchr.New(...)` constructor function.

```go
  myBatcherOptions := Opts{
		PollingInterval:    2 * time.Minute,  // check every 2 minutes for items potentially "stuck" in the batcher
		NumChecksAfterStop: 15,  // check 15 times for items left in the batcher, following a call to batchr.Batcher.Stop()
	}

  /****************************************************************************************************
  *  assuming the previously defined functions:                                                       *
  *   func packageAndDeliver(cakes []Cupcake) ...                                                     *
  *   func checkTheBatchSize(newCake Cupcake, existingCakesInCurrentBatch []Cupcake) bool ...         *
  *   func isItTimeForANewBatch(lastUpdated *time.Time) bool ...                                      *
  *****************************************************************************************************/
  myBatcher, err := batchr.New[Cupcake](
    packageAndDeliver, checkTheBatchSize, isItTimeForANewBatch,
    myBatcherOptions)
```

When using the `batchr.Opts` struct:
- specify a value for both parameters `PollingInterval` and `NumChecksAfterStop`
- omitting the `PollingInterval` parameter turns off periodic polling for the `batchr.Batcher` object (at the end of processing, some items will likely get stuck)
- omitting the `PollingInterval` parameter deactivates the `NumChecksAfterStop` parameter
