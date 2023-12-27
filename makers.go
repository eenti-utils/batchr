package batchr

import (
	"errors"
	"fmt"
	"sync"
	"time"
)


var _ Batcher[int] = (*batchMaker[int])(nil)

type batchMaker[V any] struct {
	currIdx       int
	holder        batchItems[V]
	process       BatchProcessor[V]
	isAtCap       CapacityEvaluator[V]
	checkInterval IntervalEvaluator
	itemCh        chan V
	stopRqsted    bool
	batchCount    int
	isProcessing  bool
	mu            sync.Mutex
	opts          Opts
}

func New[V any](
	p BatchProcessor[V],
	ce CapacityEvaluator[V],
	ie IntervalEvaluator,
	opts0 ...Opts,
) (Batcher[V], error) {
	var batcherOpts Opts = Opts{
		PollingInterval:    time.Second,
		NumChecksAfterStop: 3,
	}
	
	if len(opts0) > 0 {
		batcherOpts = opts0[0]
	}

	return newBatchMaker[V](
		p, ce, ie, batcherOpts,
	)
}

func newBatchMaker[V any](
	p BatchProcessor[V],
	ce CapacityEvaluator[V],
	ie IntervalEvaluator,
	batcherOpts Opts,
) (r *batchMaker[V], e error) {
	if p == nil {
		e = errors.New("nil batch processor function")
		return
	}

	if ce == nil {
		e = errors.New("nil capacity evaluator function")
		return
	}

	if ie == nil {
		e = errors.New("nil interval evaluator function")
		return
	}

	if batcherOpts.NumChecksAfterStop <0 {
		e = fmt.Errorf("invalid value - NumChecksAfterStop - [ %d ]",batcherOpts.NumChecksAfterStop)
		return 
	}

	initBatchMaker := func(maker *batchMaker[V]) {
		go func() {
			acceptingItems := true
			for acceptingItems {
				item, itemOk := <-maker.itemCh

				if itemOk {
					maker.doAddItem(item)
				} else {
					acceptingItems = false
				}
			}
		}()

		if maker.opts.PollingInterval > 0 {
			go func() {
				defer maker.closeItemsChannel()

				periodicChecking := true
				checksPerformedAfterStop := 0
				for periodicChecking {
					time.Sleep(maker.opts.PollingInterval)

					if maker.shouldProcessAfterTime() {
						maker.processBatch(trgInt)
					}

					if maker.stopRqsted {
						if checksPerformedAfterStop >= maker.opts.NumChecksAfterStop {
							return //shut it down
						}
						checksPerformedAfterStop++
					}
				}
			}()
		}
	}

	bMaker := &batchMaker[V]{
		holder:        make(batchItems[V]),
		itemCh:        make(chan V),
		process:       p,
		isAtCap:       ce,
		checkInterval: ie,
		opts: batcherOpts,
	}

	bMaker.holder.initNewItems(bMaker.currIdx)

	initBatchMaker(bMaker)

	r = bMaker
	return
}

func (m *batchMaker[V]) current() int {
	return m.currIdx
}

func (m *batchMaker[V]) next() int {
	m.currIdx++
	m.batchCount++
	return m.currIdx
}

func (m *batchMaker[V]) items() []V {
	return m.holder.fetchItems(m.current())
}

func (m *batchMaker[V]) doAddItem(item V) {

	if !m.isAtCapacity(item) {
		m.holder.add(m.current(), item)
		return
	}

	m.processBatch(trgCap)
	m.holder.add(m.current(), item)
}

func (m *batchMaker[V]) isAtCapacity(item V) bool {
	return m.isAtCap(item, m.items())
}

func (m *batchMaker[V]) processBatch(_ procTrig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isProcessing = true
	itemBatch := m.items()
	if len(itemBatch) == 0 {
		m.isProcessing = false
		return
	}
	m.process(itemBatch)
	oldIdx := m.current()
	newIdx := m.next()
	m.holder.nextItems(oldIdx, newIdx)
	m.isProcessing = false
}

func (m *batchMaker[V]) IsEmpty() bool {
	if exists, isEmpty := m.holder.existsAndEmpty(m.current()); exists {
		return isEmpty
	}
	return true
}

func (m *batchMaker[V]) shouldProcessAfterTime() (r bool) {
	if !m.IsEmpty() && !m.isProcessing {
		lastUpdated := m.holder.fetchLastUpdated(m.current())
		r = m.checkInterval(lastUpdated)
	}
	return
}

func (m *batchMaker[V]) Stop() {
	defer func() {
		m.mu.Unlock()
		m.processBatch(trgStop)
	}()

	m.mu.Lock()
	m.stopRqsted = true
}

func (m *batchMaker[V]) Add(items0 ...V) (r bool) {
	if m.stopRqsted {
		return
	}

	for _, item := range items0 {
		m.itemCh <- item
	}
	r = true
	return
}

func (m *batchMaker[V]) Size() (r int) {
	if !m.IsEmpty() {
		r = m.holder.size(m.current())
	}
	return
}

func (m *batchMaker[V]) BatchCount() int {
	return m.batchCount
}

func (m *batchMaker[V]) closeItemsChannel() {
	defer func ()  {
		recover()
	}()
	close(m.itemCh)
}