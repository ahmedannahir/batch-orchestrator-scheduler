package ExecutionStatus

type ExecutionStatus string

const (
	RUNNING   ExecutionStatus = "0"
	FAILED    ExecutionStatus = "1"
	COMPLETED ExecutionStatus = "2"
	ABORTED   ExecutionStatus = "3"
	// RUNNING   ExecutionStatus = "RUNNING"
	// FAILED    ExecutionStatus = "FAILED"
	// COMPLETED ExecutionStatus = "COMPLETED"
	// ABORTED   ExecutionStatus = "ABORTED"
)
