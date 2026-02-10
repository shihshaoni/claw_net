package attribution

import "clawnet/internal/protocol"

type FailureReason string

const (
	FailureTimeout             FailureReason = "TIMEOUT"
	FailureBudgetExceeded      FailureReason = "BUDGET_EXCEEDED"
	FailureInsufficientEvidence FailureReason = "INSUFFICIENT_EVIDENCE"
	FailureConflictingClaims   FailureReason = "CONFLICTING_CLAIMS"
	FailureToolError           FailureReason = "TOOL_ERROR"
	FailureInvalidState        FailureReason = "INVALID_STATE"
	FailureLoopDetected        FailureReason = "LOOP_DETECTED"
	FailurePolicyViolation     FailureReason = "POLICY_VIOLATION"
)

type FailureReport struct {
	RunID         string
	TaskID        string
	FailingAgent  string
	FailingIntent protocol.IntentType
	MessageID     string
	UpstreamChain []string
	Reason        FailureReason
	Explanation   string
}

func (r FailureReport) String() string {
	chain := ""
	for i, c := range r.UpstreamChain {
		if i == 0 {
			chain += c
		} else {
			chain += " -> " + c
		}
	}
	if chain == "" {
		chain = "(none)"
	}
	return "RunID=" + r.RunID +
		" TaskID=" + r.TaskID +
		" Reason=" + string(r.Reason) +
		" FailingAgent=" + r.FailingAgent +
		" FailingIntent=" + string(r.FailingIntent) +
		" MessageID=" + r.MessageID +
		" UpstreamChain=" + chain +
		" Explanation=" + r.Explanation
}
