package engine

import "fmt"

// FunctionNotFoundError indicates the requested function does not exist.
type FunctionNotFoundError struct {
	FunctionID string
}

func (e *FunctionNotFoundError) Error() string {
	return fmt.Sprintf("function not found: %s", e.FunctionID)
}

// FunctionDisabledError indicates the function is disabled.
type FunctionDisabledError struct {
	FunctionID string
}

func (e *FunctionDisabledError) Error() string {
	return fmt.Sprintf("function is disabled: %s", e.FunctionID)
}

// NoActiveVersionError indicates no active version exists for the function.
type NoActiveVersionError struct {
	FunctionID string
}

func (e *NoActiveVersionError) Error() string {
	return fmt.Sprintf("no active version found for function: %s", e.FunctionID)
}

// ExecutionRecordError indicates a failure to create/update execution record.
type ExecutionRecordError struct {
	Err error
}

func (e *ExecutionRecordError) Error() string {
	return fmt.Sprintf("execution record error: %v", e.Err)
}

func (e *ExecutionRecordError) Unwrap() error {
	return e.Err
}
