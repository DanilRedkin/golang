package crawler

import (
	"context"
	"crawler/internal/fs"
	"crawler/internal/workerpool"
	"encoding/json"
	"fmt"
	"sync"
)

// Configuration holds the configuration for the crawler, specifying the number of workers for
// file searching, processing, and accumulating tasks. The values for SearchWorkers, FileWorkers,
// and AccumulatorWorkers are critical to efficient performance and must be defined in
// every configuration.
type Configuration struct {
	SearchWorkers      int // Number of workers responsible for searching files.
	FileWorkers        int // Number of workers for processing individual files.
	AccumulatorWorkers int // Number of workers for accumulating results.
}

// Combiner is a function type that defines how to combine two values of type R into a single
// result. Combiner is not required to be thread-safe
//
// Combiner can either:
//   - Modify one of its input arguments to include the result of the other and return it,
//     or
//   - Create a new combined result based on the inputs and return it.
//
// It is assumed that type R has a neutral element (forming a monoid)
type Combiner[R any] func(current R, accum R) R

// Crawler represents a concurrent crawler implementing a map-reduce model with multiple workers
// to manage file processing, transformation, and accumulation tasks. The crawler is designed to
// handle large sets of files efficiently, assuming that all files can fit into memory
// simultaneously.
type Crawler[T, R any] interface {
	// Collect performs the full crawling operation, coordinating with the file system
	// and worker pool to process files and accumulate results. The result type R is assumed
	// to be a monoid, meaning there exists a neutral element for combination, and that
	// R supports an associative combiner operation.
	// The result of this collection process, after all reductions, is returned as type R.
	//
	// Important requirements:
	// 1. Number of workers in the Configuration is mandatory for managing workload efficiently.
	// 2. FileSystem and Accumulator must be thread-safe.
	// 3. Combiner does not need to be thread-safe.
	// 4. If an accumulator or combiner function modifies one of its arguments,
	//    it should return that modified value rather than creating a new one,
	//    or alternatively, it can create and return a new combined result.
	// 5. Context cancellation is respected across workers.
	// 6. Type T is derived by json-deserializing the file contents, and any issues in deserialization
	//    must be handled within the worker.
	// 7. The combiner function will wait for all workers to complete, ensuring no goroutine leaks
	//    occur during the process.
	Collect(
		ctx context.Context,
		fileSystem fs.FileSystem,
		root string,
		conf Configuration,
		accumulator workerpool.Accumulator[T, R],
		combiner Combiner[R],
	) (R, error)
}

type crawlerImpl[T, R any] struct{}

func New[T, R any]() *crawlerImpl[T, R] {
	return &crawlerImpl[T, R]{}
}

func (c *crawlerImpl[T, R]) Collect(
	ctx context.Context,
	fileSystem fs.FileSystem,
	root string,
	conf Configuration,
	accumulator workerpool.Accumulator[T, R],
	combiner Combiner[R],
) (R, error) {

	files := make(chan string) // Channel for paths
	var res R                  // Final accumulated result
	var errorMtx sync.Mutex    // For error sync
	var caughtError error      // A variable for error itself as we are not allowed to use buffered channels

	// Going through all files
	filePool := workerpool.New[string, string]()
	var wg sync.WaitGroup
	wg.Add(1)
	// Launching goroutine to handle the file search
	go func() {
		defer wg.Done()
		filePool.List(ctx, conf.SearchWorkers, root, c.fileSearcher(fileSystem, files, ctx, &errorMtx, &caughtError))
		close(files) // Close files channel when searching completed
	}()
	// Transforming all files using workerpool
	transformPool := workerpool.New[string, T]()
	transformedFiles := transformPool.Transform(ctx, conf.FileWorkers, files, c.fileTransformer(fileSystem, ctx, &errorMtx, &caughtError))
	// Gathering the final result using workerpool
	accumulatePool := workerpool.New[T, R]()
	accumulatedFiles := accumulatePool.Accumulate(ctx, conf.AccumulatorWorkers, transformedFiles, accumulator)

	// After accumulation, we want to combine results
	for result := range accumulatedFiles {
		select {
		case <-ctx.Done(): // Stopping when context cancellation is detected
			return res, ctx.Err()
		default:
			res = combiner(result, res) // Combining results
		}
	}
	// Waiting for all file search goroutines to complete
	wg.Wait()
	// Checking if context was canceled after completing all work
	select {
	case <-ctx.Done():
		return res, ctx.Err() // Return if the context was canceled
	default:
	}
	// Returning the encountered error if found
	if caughtError != nil {
		return res, caughtError
	}
	// If error wasn't found just nil
	return res, nil
}

// Reading directories and sending file paths to the file channel and returning subdirectories recursively
func (c *crawlerImpl[T, R]) fileSearcher(fileSystem fs.FileSystem, files chan<- string, ctx context.Context, errorMtx *sync.Mutex, caughtError *error) workerpool.Searcher[string] {
	return func(root string) []string {
		select {
		case <-ctx.Done(): // Stopping when context cancellation is detected
			return nil
		default:
		}
		// Handling any panics
		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, errorMtx, caughtError)
			}
		}()

		items, err := fileSystem.ReadDir(root) // Reading directory contents
		if err != nil {
			handleError(err, errorMtx, caughtError) // Handling any error if found
			return nil
		}

		var results []string
		for _, item := range items {
			select {
			case <-ctx.Done(): // Stopping when context cancellation is detected
				return nil
			default:
			}

			path := fileSystem.Join(root, item.Name()) // Generating full path
			if item.IsDir() {
				results = append(results, path) // Adding directories for further searching
			} else {
				select {
				case files <- path: // Sending file path to the channel for processing
				case <-ctx.Done(): // Stopping when context cancellation is detected
					return nil
				}
			}
		}
		return results // Returning the list of directories for processing recursively
	}
}

// Reading, deserializing and transforming the content of a file into type T
func (c *crawlerImpl[T, R]) fileTransformer(fileSystem fs.FileSystem, ctx context.Context, errorMtx *sync.Mutex, caughtError *error) workerpool.Transformer[string, T] {
	return func(filePath string) T {
		select {
		case <-ctx.Done(): // Stopping when context cancellation is detected
			return *new(T)
		default:
		}

		var result T
		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, errorMtx, caughtError) // Panic handling is in the separate func to avoid copypaste
			}
		}()

		file, err := fileSystem.Open(filePath) // Opening the file
		if err != nil {
			handleError(err, errorMtx, caughtError) // Error handling is in the separate func to avoid copypaste
			return result
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(&result) // Decoding file
		if err != nil {
			handleError(err, errorMtx, caughtError) // Recording any decoding errors
			return result
		}

		select {
		case <-ctx.Done(): // Stopping when context cancellation is detected
			return result
		default:
		}

		return result // Returning the transformed result
	}
}

func handleError(err error, errorMtx *sync.Mutex, caughtError *error) {
	if err == nil {
		return // When no error to handle
	}
	errorMtx.Lock() // Using mutex for thread-safeness
	defer errorMtx.Unlock()
	if *caughtError == nil {
		*caughtError = err // Recording the encountered error
	}
}

func handlePanic(r interface{}, errorMtx *sync.Mutex, caughtError *error) {
	if r == nil {
		return // When no error to handle
	}
	var err error
	switch v := r.(type) {
	case error:
		// Converting panic value to an error
		err = v // Recording the encountered error
	default:
		err = fmt.Errorf("panic: %v", r) // Creating an error if wasn't error type
	}
	handleError(err, errorMtx, caughtError) // Handling the error from panic
}
