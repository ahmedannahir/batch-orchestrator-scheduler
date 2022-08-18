package ExecutionStatus

type ExecutionStatus string

const (
	RUNNING   ExecutionStatus = "1"
	FAILED    ExecutionStatus = "2"
	COMPLETED ExecutionStatus = "3"
	ABORTED   ExecutionStatus = "4"
	// RUNNING   ExecutionStatus = "RUNNING"
	// FAILED    ExecutionStatus = "FAILED"
	// COMPLETED ExecutionStatus = "COMPLETED"
	// ABORTED   ExecutionStatus = "ABORTED"
)
