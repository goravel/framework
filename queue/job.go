package queue

import (
	"errors"
	"fmt"
	"reflect"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []contractsqueue.Job) error {
	for _, job := range jobs {
		signature := job.Signature()
		if err := ValidateTask(job.Handle); err != nil {
			return err
		}
		if _, exists := JobRegistry.Load(signature); exists {
			return errors.New("Job with signature " + signature + " already exists")
		}

		JobRegistry.Store(signature, job)
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args []contractsqueue.Arg) error {
	var err error
	defer func() {
		// Recover from panic and set err.
		if e := recover(); e != nil {
			switch e := e.(type) {
			default:
				err = errors.New("invoking handle caused a panic")
			case error:
				err = e
			case string:
				err = errors.New(e)
			}

			// TODO log the error
		}
	}()

	job, err := Get(signature)
	if err != nil {
		return err
	}
	values, err := argsToValues(args)
	if err != nil {
		return err
	}

	// Invoke the handle
	results := reflect.ValueOf(job.Handle).Call(values)

	// Handle must return at least a value
	if len(results) == 0 {
		return ErrTaskReturnsNoValue
	}

	// Last returned value
	lastResult := results[len(results)-1]

	// If the last returned value is not nil, it has to be of error type, if that
	// is not the case, return error message, otherwise propagate the handle error
	// to the caller
	if !lastResult.IsNil() {
		// check that the result implements the standard error interface,
		// if not, return ErrLastReturnValueMustBeError error
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !lastResult.Type().Implements(errorInterface) {
			return ErrLastReturnValueMustBeError
		}

		// Return the standard error
		return lastResult.Interface().(error)
	}

	return nil
}

// Get gets a registered job using its signature.
// Get 使用其签名获取已注册的作业。
func Get(signature string) (contractsqueue.Job, error) {
	value, exists := JobRegistry.Load(signature)
	if !exists {
		return nil, fmt.Errorf("job %s not found", signature)
	}
	job, ok := value.(contractsqueue.Job)
	if !ok {
		return nil, errors.New("job must implement contracts/queue/Job interface")
	}

	return job, nil
}
