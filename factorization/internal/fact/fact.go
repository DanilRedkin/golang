package fact

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
)

var (
	// ErrFactorizationCancelled is returned when the factorization process is cancelled via the done channel.
	ErrFactorizationCancelled = errors.New("cancelled")

	// ErrWriterInteraction is returned if an error occurs while interacting with the writer
	// triggering early termination.
	ErrWriterInteraction = errors.New("writer interaction")
)

// Config defines the configuration for factorization and write workers.
type Config struct {
	FactorizationWorkers int
	WriteWorkers         int
}

// Factorization interface represents a concurrent prime factorization task with configurable workers.
// Thread safety and error handling are implemented as follows:
// - The provided writer must be thread-safe to handle concurrent writes from multiple workers.
// - Output uses '\n' for newlines.
// - Factorization has a time complexity of O(sqrt(n)) per number.
// - If an error occurs while writing to the writer, early termination is triggered across all workers.
type Factorization interface {
	// Do performs factorization on a list of integers, writing the results to an io.Writer.
	// - done: a channel to signal early termination.
	// - numbers: the list of integers to factorizing.
	// - writer: the io.Writer where factorization results are output.
	// - config: optional worker configuration.
	// Returns an error if the process is cancelled or if a writer error occurs.
	Do(done <-chan struct{}, numbers []int, writer io.Writer, config ...Config) error
}

// factorizationImpl provides an implementation for the Factorization interface.
type factorizationImpl struct{}

func New() *factorizationImpl {
	return &factorizationImpl{}
}

func (f *factorizationImpl) Do(
	done <-chan struct{},
	numbers []int,
	writer io.Writer,
	config ...Config,
) error {
	// Setting config if provided or default one
	cfg := getConfig(config)
	// Checking provided config and alter if needed
	if err := checkConfig(cfg); err != nil {
		return err
	}
	// Signal for stopping all workers
	errorDone := make(chan struct{})
	var customDoneOnce sync.Once
	closeCustomDone := func() {
		customDoneOnce.Do(func() {
			close(errorDone)
		})
	}
	// Channels init
	// taskChannel for each number in list
	// resultChannel for each number factorization  result
	// errorChannel for error

	taskChannel := make(chan int, cfg.FactorizationWorkers)
	resultChannel := make(chan string, len(numbers))
	errorChannel := make(chan error, 1)
	// Creating wait groups
	var factorWg sync.WaitGroup
	var writeWg sync.WaitGroup
	// Start them working with the calculations of factors
	for i := 0; i < cfg.FactorizationWorkers; i++ {
		factorWg.Add(1)
		go startFactorizationWorker(done, errorDone, taskChannel, resultChannel, &factorWg)
	}
	// Start writer workers
	for i := 0; i < cfg.WriteWorkers; i++ {
		writeWg.Add(1)
		go startWriterWorker(done, errorDone, resultChannel, closeCustomDone, errorChannel, writer, &writeWg)
	}

	// Providing numbers to the task channel
	go func() {
		defer close(taskChannel)
		for _, num := range numbers {
			select {
			case <-done:
				closeCustomDone()
				return
			case <-errorDone:
				return
			case taskChannel <- num:
			}
		}
	}()

	// Waiting for all workers to complete
	factorWg.Wait()
	// Closing channel to signal writing is done
	close(resultChannel)
	writeWg.Wait()

	// Checking if an error in writer
	select {
	case err := <-errorChannel:
		closeCustomDone()
		return err
	case <-done:
		closeCustomDone()
		return ErrFactorizationCancelled
	default:
		closeCustomDone()
		return nil
	}
}
func startFactorizationWorker(
	done <-chan struct{},
	errorDone <-chan struct{},
	taskChannel <-chan int,
	resultChannel chan<- string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for num := range taskChannel {
		select {
		case <-done:
			return
		case <-errorDone:
			return
		default:
			factors := factorizing(num)
			result := formating(num, factors)
			select {
			case <-done:
				return
			case <-errorDone:
				return
			case resultChannel <- result:
			}
		}
	}
}

func startWriterWorker(
	done <-chan struct{},
	errorDone chan struct{},
	resultChannel <-chan string,
	closeCustomDone func(),
	errorChannel chan<- error,
	writer io.Writer,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for result := range resultChannel {
		select {
		case <-done:
			return
		case <-errorDone:
			return
		default:
			if _, err := fmt.Fprintln(writer, result); err != nil {
				select {
				case errorChannel <- fmt.Errorf("%w: %w", ErrWriterInteraction, err):
					closeCustomDone()
				default:
				}
				return
			}
		}
	}
}

func factorizing(n int) []int {
	var result []int
	if n == 0 || n == 1 || n == -1 {
		return []int{n}
	}
	if n < 0 {
		result = append(result, -1)
		n = -n
	}
	for n%2 == 0 {
		result = append(result, 2)
		n /= 2
	}
	for i := 3; i*i <= n; i += 2 {
		for n%i == 0 {
			result = append(result, i)
			n /= i
		}
	}
	if n > 1 {
		result = append(result, n)
	}
	return result
}

// Formating the factors into a string
func formating(num int, factors []int) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d = ", num))
	for i, factor := range factors {
		if i > 0 {
			builder.WriteString(" * ")
		}
		builder.WriteString(fmt.Sprintf("%d", factor))
	}
	return builder.String()
}

// Getting the configuration to use, set to defaults if necessary
func getConfig(config []Config) Config {
	if len(config) > 0 {
		return config[0]
	}
	return Config{
		FactorizationWorkers: runtime.NumCPU(),
		WriteWorkers:         runtime.NumCPU(),
	}
}

// Validating configuration
func checkConfig(cfg Config) error {
	if cfg.FactorizationWorkers <= 0 || cfg.WriteWorkers <= 0 {
		return errors.New("invalid configuration: worker counts cannot be negative")
	}
	return nil
}
