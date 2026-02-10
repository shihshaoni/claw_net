package task

type State string

const (
	StateInit      State = "INIT"
	StatePlanning  State = "PLANNING"
	StateExecuting State = "EXECUTING"
	StateVerifying State = "VERIFYING"
	StateSuccess   State = "SUCCESS"
	StateFailed    State = "FAILED"
)

type Task struct {
	TaskID string
	State  State
	Round  int
}

func New(taskID string) *Task {
	return &Task{TaskID: taskID, State: StateInit, Round: 0}
}
