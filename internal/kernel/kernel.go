package kernel

import (
	"fmt"
	"time"

	"clawnet/internal/attribution"
	"clawnet/internal/constraints"
	"clawnet/internal/eventstore"
	"clawnet/internal/protocol"
	"clawnet/internal/task"
)

type Kernel struct {
	store       eventstore.Store
	constraints constraints.ConstraintSet
	task        *task.Task

	runID string

	// minimal indices for attribution v0
	lastClaimMsgID string
	lastClaimAgent string
	lastClaimRefs  []string
}

func New(runID string, taskID string, store eventstore.Store, c constraints.ConstraintSet) *Kernel {
	return &Kernel{
		store:       store,
		constraints: c,
		task:        task.New(taskID),
		runID:       runID,
	}
}

func (k *Kernel) RunID() string  { return k.runID }
func (k *Kernel) TaskID() string { return k.task.TaskID }
func (k *Kernel) State() task.State { return k.task.State }

func (k *Kernel) EmitMessage(m protocol.Message) error {
	_, err := k.store.Append(k.runID, eventstore.Event{
		EventID:   m.MessageID,
		RunID:     k.runID,
		Type:      eventstore.EventMessageSent,
		Actor:     m.FromAgent,
		Data:      m,
		Timestamp: time.Now(),
	})
	if err != nil {
		return err
	}

	// very small "semantic" bookkeeping for Phase 1 demo
	switch m.Intent {
	case protocol.IntentClaim:
		k.lastClaimMsgID = m.MessageID
		k.lastClaimAgent = m.FromAgent
		k.lastClaimRefs = m.Refs
	case protocol.IntentDecision:
		// state transition happens only on DECISION
		if err := k.task.Apply(m.Intent); err != nil {
			_, _ = k.store.Append(k.runID, eventstore.Event{
				EventID:   "err-" + m.MessageID,
				RunID:     k.runID,
				Type:      eventstore.EventErrorRaised,
				Actor:     m.FromAgent,
				Data:      err.Error(),
				Timestamp: time.Now(),
			})
			return err
		}
		_, _ = k.store.Append(k.runID, eventstore.Event{
			EventID:   "st-" + m.MessageID,
			RunID:     k.runID,
			Type:      eventstore.EventStateTransition,
			Actor:     m.FromAgent,
			Data:      fmt.Sprintf("state=%s round=%d", k.task.State, k.task.Round),
			Timestamp: time.Now(),
		})
	}

	return k.checkConstraints()
}

func (k *Kernel) checkConstraints() error {
	if k.task.Round > k.constraints.MaxRounds {
		return fmt.Errorf("constraint violated: rounds=%d > max=%d", k.task.Round, k.constraints.MaxRounds)
	}
	events, _ := k.store.List(k.runID)
	if len(events) > k.constraints.MaxMessages {
		return fmt.Errorf("constraint violated: messages=%d > max=%d", len(events), k.constraints.MaxMessages)
	}
	return nil
}

// Finalize creates a FailureReport when the run ends in FAILED or when verifier challenges insufficient evidence.
func (k *Kernel) FinalizeInsufficientEvidence(challengerAgent string, challengeMsgID string) (attribution.FailureReport, error) {
	return attribution.FailureReport{
		RunID:         k.runID,
		TaskID:        k.task.TaskID,
		FailingAgent:  k.lastClaimAgent,
		FailingIntent: protocol.IntentClaim,
		MessageID:     k.lastClaimMsgID,
		UpstreamChain: []string{k.lastClaimMsgID, challengeMsgID},
		Reason:        attribution.FailureInsufficientEvidence,
		Explanation:   "Verifier challenged the latest CLAIM due to missing or weak EVIDENCE references (v0 heuristic).",
	}, nil
}
