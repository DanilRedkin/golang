package workerpool

import (
	"context"
	"sync"
)

// Accumulator is a function type used to aggregate values of type T into a result of type R.
// It must be thread-safe, as multiple goroutines will access the accumulator function concurrently.
// Each worker will produce intermediate results, which are combined with an initial or
// accumulated value.
type Accumulator[T, R any] func(current T, accum R) R

// Transformer is a function type used to transform an element of type T to another type R.
// The function is invoked concurrently by multiple workers, and therefore must be thread-safe
// to ensure data integrity when accessed across multiple goroutines.
// Each worker independently applies the transformer to its own subset of data, and although
// no shared state is expected, the transformer must handle any internal state in a thread-safe
// manner if present.
type Transformer[T, R any] func(current T) R

// Searcher is a function type for exploring data in a hierarchical manner.
// Each call to Searcher takes a parent element of type T and returns a slice of T representing
// its child elements. Since multiple goroutines may call Searcher concurrently, it must be
// thread-safe to ensure consistent results during recursive  exploration.
//
// Important considerations:
//  1. Searcher should be designed to avoid race conditions, particularly if it captures external
//     variables in closures.
//  2. The calling function must handle any state or values in closures, ensuring that
//     captured variables remain consistent throughout recursive or hierarchical search paths.
type Searcher[T any] func(parent T) []T

// Pool is the primary interface for managing worker pools, with support for three main
// operations: Transform, Accumulate, and List. Each operation takes an input channel, applies
// a transformation, accumulation, or list expansion, and returns the respective output.
type Pool[T, R any] interface {
	// Transform applies a transformer function to each item received from the input channel,
	// with results sent to the output channel. Transform operates concurrently, utilizing the
	// specified number of workers. The number of workers must be explicitly defined in the
	// configuration for this function to handle expected workloads effectively.
	// Since multiple workers may call the transformer function concurrently, it must be
	// thread-safe to prevent race conditions or unexpected results when handling shared or
	// internal state. Each worker independently applies the transformer function to its own
	// data subset.
	Transform(ctx context.Context, workers int, input <-chan T, transformer Transformer[T, R]) <-chan R

	// Accumulate applies an accumulator function to the items received from the input channel,
	// with results accumulated and sent to the output channel. The accumulator function must
	// be thread-safe, as multiple workers concurrently update the accumulated result.
	// The output channel will contain intermediate accumulated results as R
	Accumulate(ctx context.Context, workers int, input <-chan T, accumulator Accumulator[T, R]) <-chan R

	// List expands elements based on a searcher function, starting
	// from the given element. The searcher function finds child elements for each parent,
	// allowing exploration in a tree-like structure.
	// The number of workers should be configured based on the workload, ensuring each worker
	// independently processes assigned elements.
	List(ctx context.Context, workers int, start T, searcher Searcher[T])
}

type poolImpl[T, R any] struct{}

func New[T, R any]() *poolImpl[T, R] {
	return &poolImpl[T, R]{}
}

func (p *poolImpl[T, R]) Accumulate(ctx context.Context, workers int, input <-chan T, accumulator Accumulator[T, R]) <-chan R {
	output := make(chan R)
	var wg sync.WaitGroup // WaitGroup for synchronizing workers
	// Worker function to process items concurrently and accumulate the results
	worker := func() {
		defer wg.Done()
		var res R // accumulator result
		for {
			select {
			case <-ctx.Done(): // Stopping if context is canceled
				return
			case item, ok := <-input:
				if !ok { // When input is closed
					select {
					case output <- res: // Sending the final result
					case <-ctx.Done(): // Stopping if context is canceled
					}
					return
				}
				res = accumulator(item, res) // Accumulating the result
			}
		}
	}
	// Launching workers goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}
	// Waiting for all workers to complete and then close the output channel
	go func() {
		defer close(output) // Closing channel after completing job
		wg.Wait()           // Waiting for all workers to complete
	}()

	return output // Returning the output channel with accumulated results
}

func (p *poolImpl[T, R]) List(ctx context.Context, workers int, start T, searcher Searcher[T]) {
	var processLayer func(layer []T)
	// Recursively handling the exploration of elements
	processLayer = func(cur []T) {
		if len(cur) == 0 {
			return // If there are no more elements to process
		}

		queue := make(chan T) // Holding elements to be further processed for workers
		next := make([]T, 0)  // Storing the next layer

		var layerMutex sync.Mutex
		var wg sync.WaitGroup // WaitGroup to synchronize workers

		for worker := 0; worker < workers; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done(): // Stopping if context is canceled
						return
					case task, ok := <-queue:
						if !ok {
							return
						}
						descendants := searcher(task) // Finding child elements
						layerMutex.Lock()
						next = append(next, descendants...)
						layerMutex.Unlock()
					}
				}
			}()
		}
		// Starting a goroutine to provide the queue with elements to be processed by workers
		go func() {
			defer close(queue) // Closing when done
			for _, item := range cur {
				select {
				case <-ctx.Done(): // Stopping if context is canceled
					return
				case queue <- item: // Sending item to the queue for further processing
				}
			}
		}()

		wg.Wait()          // Waiting for all workers to complete cur layer
		processLayer(next) // Recursively processing next layer
	}

	processLayer([]T{start}) // Starting recursive processing from the root
}

func (p *poolImpl[T, R]) Transform(
	ctx context.Context,
	workers int,
	input <-chan T,
	transformer Transformer[T, R],
) <-chan R {

	output := make(chan R) // Channel for sending transformed results
	var wg sync.WaitGroup  // WaitGroup to synchronize workers
	// Worker function to apply transformer
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done(): // Stopping if context is canceled
				return
			case item, ok := <-input:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case output <- transformer(item):
				}
			}
		}
	}
	// Launching workers goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}
	// Waiting for all workers to complete and then close the output channel
	go func() {
		defer close(output)
		wg.Wait()
	}()

	return output
}
