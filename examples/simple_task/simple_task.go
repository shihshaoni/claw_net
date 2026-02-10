package simple_task

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"clawnet/internal/constraints"
	"clawnet/internal/eventstore"
	"clawnet/internal/kernel"
	"clawnet/internal/protocol"
	"clawnet/internal/attribution"
)

func uuidLike() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Run executes a minimal 3-agent scenario:
// - Planner issues decisions to progress states
// - Worker emits a CLAIM without sufficient EVIDENCE
// - Verifier CHALLENGE triggers an attribution report
func Run() (attribution.FailureReport, error) {
	runID := uuidLike()
	taskID := "simple_task_v0"

	store := eventstore.NewInMemory()
	k := kernel.New(runID, taskID, store, constraints.Default())

	// Planner: INIT -> PLANNING
	_ = k.EmitMessage(protocol.Message{
		MessageID: uuidLike(),
		RunID: runID,
		TaskID: taskID,
		FromAgent: "planner",
		ToAgents: []string{"kernel"},
		Intent: protocol.IntentDecision,
		Payload: map[string]any{"transition": "INIT->PLANNING"},
		Timestamp: time.Now(),
	})

	// Planner: PLANNING -> EXECUTING
	_ = k.EmitMessage(protocol.Message{
		MessageID: uuidLike(),
		RunID: runID,
		TaskID: taskID,
		FromAgent: "planner",
		ToAgents: []string{"kernel"},
		Intent: protocol.IntentDecision,
		Payload: map[string]any{"transition": "PLANNING->EXECUTING"},
		Timestamp: time.Now(),
	})

	// Worker: emits CLAIM with no evidence refs (intentional)
	claimID := uuidLike()
	_ = k.EmitMessage(protocol.Message{
		MessageID: claimID,
		RunID: runID,
		TaskID: taskID,
		FromAgent: "worker",
		ToAgents: []string{"kernel"},
		Intent: protocol.IntentClaim,
		Payload: map[string]any{"claim": "Result is correct."},
		Confidence: 0.62,
		Refs: []string{}, // <-- missing evidence on purpose
		Timestamp: time.Now(),
	})

	// Planner: EXECUTING -> VERIFYING (decision)
	_ = k.EmitMessage(protocol.Message{
		MessageID: uuidLike(),
		RunID: runID,
		TaskID: taskID,
		FromAgent: "planner",
		ToAgents: []string{"kernel"},
		Intent: protocol.IntentDecision,
		Payload: map[string]any{"transition": "EXECUTING->VERIFYING"},
		Timestamp: time.Now(),
	})

	// Verifier: CHALLENGE the claim
	challengeID := uuidLike()
	_ = k.EmitMessage(protocol.Message{
		MessageID: challengeID,
		RunID: runID,
		TaskID: taskID,
		FromAgent: "verifier",
		ToAgents: []string{"kernel"},
		Intent: protocol.IntentChallenge,
		Payload: map[string]any{"challenge": "No evidence provided for CLAIM."},
		Refs: []string{claimID},
		Timestamp: time.Now(),
	})

	// v0: we convert this into a failure report
	return k.FinalizeInsufficientEvidence("verifier", challengeID)
}
