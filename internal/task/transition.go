package task

import (
	"fmt"

	"clawnet/internal/protocol"
)

type Transition struct {
	From   State
	To     State
	Intent protocol.IntentType
}

var allowed = map[State]map[protocol.IntentType]State{
	StateInit: {
		protocol.IntentDecision: StatePlanning,
	},
	StatePlanning: {
		protocol.IntentDecision: StateExecuting,
	},
	StateExecuting: {
		protocol.IntentDecision: StateVerifying,
		protocol.IntentChallenge: StateExecuting, // challenge can keep in executing (v0)
	},
	StateVerifying: {
		protocol.IntentDecision: StateSuccess,
		protocol.IntentChallenge: StateExecuting, // verifier can push back
	},
}

func (t *Task) Apply(intent protocol.IntentType) error {
	nextByIntent, ok := allowed[t.State]
	if !ok {
		return fmt.Errorf("no transitions from state=%s", t.State)
	}
	next, ok := nextByIntent[intent]
	if !ok {
		return fmt.Errorf("invalid transition from state=%s via intent=%s", t.State, intent)
	}
	t.State = next
	t.Round++
	return nil
}
