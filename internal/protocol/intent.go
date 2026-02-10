package protocol

type IntentType string

const (
	IntentRequest   IntentType = "REQUEST"
	IntentResponse  IntentType = "RESPONSE"
	IntentClaim     IntentType = "CLAIM"
	IntentEvidence  IntentType = "EVIDENCE"
	IntentChallenge IntentType = "CHALLENGE"
	IntentStatus    IntentType = "STATUS"
	IntentDecision  IntentType = "DECISION"
)
