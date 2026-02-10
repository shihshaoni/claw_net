# Benchmark Task v0: Evidence-First Decision Under Constraints

This benchmark is designed to **force** the interaction patterns ClawNet cares about:

- CLAIM must be backed by EVIDENCE
- Verifier must be able to CHALLENGE and push back
- Planner must DECIDE state transitions
- The run must end with an attributable failure if evidence is missing

## Task: "Change Impact Verification"

**Scenario**
A system proposes a change that could impact revenue. The Worker claims the change is safe.

**Goal**
The system must reach a final DECISION to proceed **only if** evidence is provided.

## Required interaction structure (minimum)

1. Planner: DECISION INIT→PLANNING
2. Planner: DECISION PLANNING→EXECUTING
3. Worker: CLAIM "Change is safe"
4. Worker: EVIDENCE references (must include >= 1 artifact/ref)
5. Planner: DECISION EXECUTING→VERIFYING
6. Verifier: CHALLENGE if evidence missing/weak
7. Planner: DECISION VERIFYING→SUCCESS or Verifier triggers failure

## Pass condition

- System reaches SUCCESS with evidence coverage >= 1 and verifier accepts.

## Fail conditions

- INSUFFICIENT_EVIDENCE: Worker CLAIM has empty/invalid refs, verifier challenges.
- CONFLICTING_CLAIMS: multiple workers disagree without resolution.
- TIMEOUT / BUDGET_EXCEEDED: constraints violated.

## Why this benchmark?

It is small, deterministic, and maps directly to **engineering correctness**:
- pre-release validation
- responsibility chains
- reproducibility for postmortems

Next: implement this benchmark as a runnable scenario similar to `simple_task`.
