# ClawNet

> Last updated: 2026-02-28 21:49 (UTC+8)

**ClawNet is not an agent framework.**

ClawNet is a **multi-agent interaction kernel** that makes agent interaction:

- **Observable** — interactions are typed intents, not raw text
- **Replayable** — every run is event-sourced and deterministic
- **Attributable** — failures produce responsibility chains

When a multi-agent system fails, ClawNet can precisely identify:

- **Which agent**
- **At which interaction**
- **Violated which invariant**

ClawNet does not make agents smarter.
It makes agent interaction **controllable, accountable, and evolvable**.

---

## Why ClawNet?

Multi-agent systems are increasingly used to solve complex tasks, but most existing frameworks fail at a fundamental engineering requirement: **when the system fails, we cannot explain why in an actionable way.**

Today, multi-agent failures are typically described as:
- "the model hallucinated"
- "agents misunderstood each other"
- "the prompt needs tuning"

These explanations are not acceptable in production engineering, because they provide:
- no reproducible root cause
- no regression test
- no concrete corrective action

The core problem is not intelligence — it is **observability**. Multi-agent systems are distributed decision systems:
- state is fragmented across agents
- decisions are interdependent
- errors emerge late and indirectly

Yet most frameworks treat agent interaction as unstructured chat logs. This is equivalent to running a distributed system without tracing, replay, or postmortems.

Every mature engineering domain went through the same transition:

| Domain | Before | After |
|--------|--------|-------|
| Web | ad-hoc requests | HTTP specification |
| Microservices | logs only | tracing & spans |
| Distributed systems | best effort | consensus & invariants |
| **Multi-agent** | **prompt chaining** | **interaction kernel** |

**ClawNet does not add complexity — it exposes the complexity that already exists.**

---

## Quick Start

```bash
go run ./cmd/clawnetd
```

You should see:
- a small 3-agent run (`Planner`, `Worker`, `Verifier`)
- an intentional failure (`INSUFFICIENT_EVIDENCE`)
- a generated **FailureReport** with a responsibility chain

### Example Output

```
ClawNet - Phase 1 Engineering Skeleton
Running examples/simple_task ...
-----
Run finished in: 109.292us

FailureReport:
  RunID          = 4c2add4c5332f0a6...
  TaskID         = simple_task_v0
  Reason         = INSUFFICIENT_EVIDENCE
  FailingAgent   = worker
  FailingIntent  = CLAIM
  MessageID      = de620323edd9b6e1...
  UpstreamChain  = de620323... -> e23e2bba...
  Explanation    = Verifier challenged the latest CLAIM due to
                   missing or weak EVIDENCE references (v0 heuristic).
```

This tells you: **Worker** made a **CLAIM** without providing evidence refs. **Verifier** challenged it. ClawNet traced the responsibility chain back to the exact message where the failure originated.

---

## Core Concepts

### Intent-Based Messaging

ClawNet replaces free-form chat with **typed intents**. Every agent interaction must declare its semantic purpose:

```go
// internal/protocol/intent.go
type IntentType string

const (
    IntentRequest   IntentType = "REQUEST"    // Ask another agent to do something
    IntentResponse  IntentType = "RESPONSE"   // Answer a request
    IntentClaim     IntentType = "CLAIM"      // Assert a result (requires evidence)
    IntentEvidence  IntentType = "EVIDENCE"   // Provide supporting data for a claim
    IntentChallenge IntentType = "CHALLENGE"  // Dispute a claim
    IntentStatus    IntentType = "STATUS"     // Report progress
    IntentDecision  IntentType = "DECISION"   // Advance the task state machine
)
```

Each message carries structured metadata for tracing and attribution:

```go
// internal/protocol/message.go
type Message struct {
    MessageID string        // Unique ID for this interaction
    RunID     string        // Groups messages into a single run
    TaskID    string        // The task being worked on

    FromAgent string        // Who sent this
    ToAgents  []string      // Who receives this

    Intent     IntentType   // Semantic purpose (not free text)
    Payload    any          // Content (variable, doesn't affect system layer)
    Confidence float64      // Agent's self-assessed confidence
    Refs       []string     // Responsibility chain — links to prior messages

    IdempotencyKey string
    Timestamp      time.Time
}
```

The moat is `Intent + Refs`, not `Payload`. Intents make interactions computable; Refs make failures traceable.

### How ClawNet Differs from OpenTelemetry

Readers familiar with distributed tracing may ask: *"Why not just use OpenTelemetry?"*

| | OpenTelemetry | ClawNet |
|--|---------------|---------|
| **Design for** | Passive observation of existing systems | Active enforcement of interaction structure |
| **Data model** | Generic spans and events | Typed intents with semantic meaning |
| **Agent awareness** | None — treats agents as black boxes | First-class — agents have roles, constraints, responsibility |
| **Failure analysis** | "Span X took 5s" | "Agent Y made CLAIM Z without evidence, violating invariant" |
| **Replay** | Not supported | Event-sourced strict/soft replay |
| **Attribution** | Latency & error rates | Causal responsibility chains |

OpenTelemetry answers: *"What happened and how long did it take?"*
ClawNet answers: *"Which agent, at which interaction, violated which invariant, and why?"*

ClawNet can export to OTel-compatible backends for visualization, but its core value is the **typed interaction protocol** and **causal attribution** that OTel was never designed to provide.

---

## Repo Layout

```
clawnet/
├── cmd/clawnetd/main.go              # Entry point
├── docs/
│   ├── ClawNet_Master_Design.md      # Comprehensive design spec
│   ├── ClawNet_Engineering_Architecture_Map.md
│   └── ClawNet_Whitepaper.md
├── examples/
│   ├── simple_task/simple_task.go     # Minimal 3-agent failure scenario
│   └── benchmark_task/README.md      # Future benchmark spec
└── internal/
    ├── protocol/                     # Intent types & message schema
    │   ├── intent.go                 # 7 IntentTypes (REQUEST, RESPONSE, CLAIM, ...)
    │   └── message.go                # Message struct with refs for attribution
    ├── eventstore/                   # Append-only event log
    │   ├── event.go                  # Event struct & 5 EventTypes
    │   └── store.go                  # InMemoryStore (v0, thread-safe)
    ├── task/                         # Task state machine
    │   ├── state.go                  # 6 states: INIT→PLANNING→EXECUTING→VERIFYING→SUCCESS/FAILED
    │   └── transition.go             # FSM validation & Apply()
    ├── constraints/                  # Resource constraints (rounds, messages, budget)
    │   └── constraints.go
    ├── kernel/                       # Central orchestration
    │   └── kernel.go                 # Message routing, state transitions, event emission
    └── attribution/                  # Failure analysis
        └── failure.go                # FailureReport & 8 FailureReasons
```

---

## Phase 1 Database Schema (Target)

### traces

```sql
CREATE TABLE traces (
  trace_id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ DEFAULT now(),
  total_token_in BIGINT DEFAULT 0,
  total_token_out BIGINT DEFAULT 0,
  total_cost_usd NUMERIC(18,8) DEFAULT 0,
  total_latency_ms BIGINT DEFAULT 0,
  max_depth INT DEFAULT 0,
  fanout_max INT DEFAULT 0,
  cycle_count INT DEFAULT 0
);
```

### interactions

```sql
CREATE TABLE interactions (
  interaction_id UUID PRIMARY KEY,
  trace_id UUID REFERENCES traces(trace_id),
  parent_interaction_id UUID,
  from_type TEXT,
  from_id TEXT,
  to_type TEXT,
  to_id TEXT,
  kind TEXT,
  status TEXT,
  ts_start TIMESTAMPTZ,
  ts_end TIMESTAMPTZ,
  token_in BIGINT DEFAULT 0,
  token_out BIGINT DEFAULT 0,
  cost_usd NUMERIC(18,8) DEFAULT 0,
  latency_ms BIGINT DEFAULT 0,
  tags JSONB DEFAULT '{}',
  payload_id UUID
);
```

---

## Go SDK Interfaces (Target)

### Tracer

```go
type Tracer struct {
  exporter Exporter
}

func (t *Tracer) StartTrace(ctx context.Context, name string, tags map[string]any) (context.Context, *Trace)
func (t *Tracer) StartSpan(ctx context.Context, kind string, from, to Endpoint, tags map[string]any) (context.Context, *Span)
```

### Span

```go
type Span struct {
  InteractionID string
  TraceID string
  ParentID *string
}

func (s *Span) SetTokens(in, out int64)
func (s *Span) SetCostUSD(cost float64)
func (s *Span) SetStatus(status string)
func (s *Span) End(err error)
```

### Exporter

```go
type Exporter interface {
  EmitInteraction(interaction Interaction) error
  EmitBatch(interactions []Interaction) error
  Close() error
}
```

---

## Collector API (Target)

### POST /ingest

Accepts a batch of interactions.

### GET /traces/{trace_id}

Returns:
- full interaction list
- computed metrics (cost, depth, fan-out, cycles)

---

# Execution Roadmap

## Execution Order: 4 → 1 → 3 → 2

| Order | Phase | Name | Status |
|-------|-------|------|--------|
| 1st | **Phase 4** | Market Validation | `NOT STARTED` |
| 2nd | **Phase 1** | Interaction Trace MVP | `IN PROGRESS` — Engineering Skeleton Complete |
| 3rd | **Phase 3** | Formal Specification / Whitepaper | `NOT STARTED` |
| 4th | **Phase 2** | Fundraising Narrative | `NOT STARTED` |

---

## Why This Order?

1. **Phase 4 first** — Validate pain before building infrastructure. Confirm that debugging LLM workflows is a real, unsolved problem.
2. **Phase 1 second** — Build the minimal working product that solves the validated pain.
3. **Phase 3 third** — Formalize what we've learned into a specification, giving ClawNet academic and engineering credibility.
4. **Phase 2 last** — With validation + product + spec in hand, fundraise with evidence, not promises.

---

# Phase 4 — Market Validation (First)

> **Goal:** Prove that multi-agent observability is a real, paying pain point before writing more infrastructure code.

### 4.1 Problem Discovery Interviews

- [ ] Identify 15-20 teams actively building multi-agent / LLM-orchestration systems
- [ ] Conduct structured interviews (30 min each) focusing on:
  - How do they debug when an agent chain fails?
  - What tools do they use today? (logs? LangSmith? custom?)
  - How much time/money is lost on undiagnosable failures?
- [ ] Document recurring themes and pain severity scores

### 4.2 Pain Point Prioritization

- [ ] Classify discovered pains into categories:
  - **Behavior observability** — "I can't see what happened"
  - **Loop / fan-out detection** — "Agents run in circles and burn tokens"
  - **Cost attribution** — "I don't know which agent costs the most"
  - **Failure traceability** — "I can't reproduce the bug"
  - **Governance / compliance** — "I need audit trails for AI decisions"
- [ ] Rank by frequency and severity
- [ ] Validate assumption: *"The first real pain is behavior observability, not multi-agent legality"*

### 4.3 Competitive Landscape Analysis

- [ ] Map existing solutions (LangSmith, Arize, Helicone, OpenTelemetry, custom solutions)
- [ ] Identify gaps: what do they NOT solve for multi-agent interaction?
- [ ] Define ClawNet's unique wedge: interaction-level tracing with causal attribution

### 4.4 Design Partner Recruitment

- [ ] From interviews, identify 3-5 teams willing to be early design partners
- [ ] Define success criteria with each partner (what would make them use/pay for ClawNet?)
- [ ] Establish feedback loop for Phase 1 MVP testing

### Deliverables

- Problem Discovery Report (interview summaries + patterns)
- Pain Point Matrix (ranked by frequency x severity)
- Competitive Gap Analysis
- List of 3-5 design partners with success criteria

---

# Phase 1 — Interaction Trace MVP (Second)

> **Goal:** Build a minimal, working interaction trace system that records, reconstructs, and analyzes multi-agent interactions.

### Current Status: Engineering Skeleton Complete

The following components are **fully implemented** (~459 lines of Go):

| Component | Status | Description |
|-----------|--------|-------------|
| `protocol/intent.go` | Done | 7 IntentTypes (REQUEST, RESPONSE, CLAIM, EVIDENCE, CHALLENGE, STATUS, DECISION) |
| `protocol/message.go` | Done | Message schema with refs for attribution chain |
| `eventstore/event.go` | Done | Event struct & 5 EventTypes |
| `eventstore/store.go` | Done | InMemoryStore (append-only, thread-safe) |
| `task/state.go` | Done | Task struct with 6 states |
| `task/transition.go` | Done | FSM validation with allowed transitions |
| `constraints/constraints.go` | Done | ConstraintSet (rounds, messages, budget, deadline) |
| `kernel/kernel.go` | Partial | Message routing + state transitions (only DECISION handled) |
| `attribution/failure.go` | Partial | FailureReport (only INSUFFICIENT_EVIDENCE implemented) |
| `examples/simple_task` | Done | Working 3-agent demo scenario |

### 1.1 Complete the Interaction Kernel

- [ ] Handle all intent types in kernel (not just DECISION)
- [ ] Emit `MESSAGE_RECEIVED` events (currently defined but unused)
- [ ] Integrate `BUDGET_UPDATED` event emission
- [ ] Enforce budget constraint in `checkConstraints()` (currently tracked but not enforced)
- [ ] Implement all 8 FailureReasons in attribution:
  - [ ] `TIMEOUT` — deadline exceeded
  - [ ] `BUDGET_EXCEEDED` — cost limit hit
  - [ ] `CONFLICTING_CLAIMS` — agents disagree on facts
  - [ ] `TOOL_ERROR` — external tool failure
  - [ ] `INVALID_STATE` — illegal state transition attempted
  - [ ] `LOOP_DETECTED` — circular interaction pattern
  - [ ] `POLICY_VIOLATION` — governance rule broken

### 1.2 Interaction Recording & Tracing

- [ ] Implement `Tracer` SDK (StartTrace / StartSpan API)
- [ ] Implement `Span` with token, cost, and status tracking
- [ ] Implement `Exporter` interface with at least two backends:
  - [ ] `StdoutExporter` — for local development
  - [ ] `HTTPExporter` — for collector ingestion
- [ ] Record LLM calls and tool calls as spans
- [ ] Reconstruct parent-child relationships (interaction graph)

### 1.3 Collector Service

- [ ] Implement `POST /ingest` endpoint (accept interaction batches)
- [ ] Implement `GET /traces/{trace_id}` endpoint (return full trace + metrics)
- [ ] Compute trace-level metrics:
  - [ ] Total tokens (in/out)
  - [ ] Total cost (USD)
  - [ ] Total latency
  - [ ] Max interaction depth
  - [ ] Max fan-out
  - [ ] Cycle count

### 1.4 Persistence Layer

- [ ] Replace InMemoryStore with Postgres-backed store
- [ ] Implement `traces` table (schema above)
- [ ] Implement `interactions` table (schema above)
- [ ] Add database migrations
- [ ] Maintain backward compatibility with InMemoryStore for testing

### 1.5 Pattern Detection

- [ ] Detect loops (agent A → B → A cycles)
- [ ] Detect excessive fan-out (one agent spawning too many parallel interactions)
- [ ] Cost attribution per agent / per interaction path
- [ ] Alert / flag on anomalous patterns

### 1.6 Replay System

- [ ] Implement **strict replay** — event-only, fully deterministic
- [ ] Implement **soft replay** — replay structure with live LLM calls
- [ ] Replay CLI: `clawnet replay <trace_id> [--mode strict|soft]`

### 1.7 Additional Examples & Testing

- [ ] Implement `benchmark_task` example (multi-round, multi-agent stress test)
- [ ] Add unit tests for all packages
- [ ] Add integration test: full trace → ingest → query round-trip
- [ ] Add failure scenario tests for each FailureReason

### Deliverables

- Working Tracer SDK (Go)
- Collector service with REST API
- Postgres persistence
- Loop / fan-out / cost detection
- Replay (strict + soft)
- Comprehensive test suite

---

# Phase 3 — Formal Specification / Whitepaper (Third)

> **Goal:** Formalize the interaction model into a rigorous specification. This gives ClawNet academic credibility and makes the protocol implementable by third parties.

### 3.1 Interaction Model Specification

- [ ] Define formal grammar for interaction types (intent taxonomy)
- [ ] Specify message envelope format (fields, types, constraints)
- [ ] Define interaction graph as a DAG with typed edges
- [ ] Specify causal ordering guarantees
- [ ] Define idempotency semantics

### 3.2 Propagation Model

- [ ] Formalize how context propagates through interaction chains
- [ ] Define trace context format (analogous to W3C Trace Context for agents)
- [ ] Specify baggage propagation rules
- [ ] Define sampling strategies for high-volume systems

### 3.3 Pattern Taxonomy

- [ ] Classify interaction patterns:
  - **Sequential** — A → B → C
  - **Fan-out** — A → {B, C, D}
  - **Fan-in** — {B, C, D} → A
  - **Loop** — A → B → A (with termination conditions)
  - **Delegation** — A → B (with authority transfer)
  - **Challenge-Response** — A claims, B verifies
- [ ] Define pattern detection algorithms
- [ ] Specify pattern-level invariants (e.g., "fan-out must converge")

### 3.4 Governance Abstraction

- [ ] Define policy language for interaction constraints
- [ ] Specify role-based interaction permissions
- [ ] Define audit trail requirements
- [ ] Specify compliance reporting format

### 3.5 Whitepaper

- [ ] Write formal whitepaper covering:
  - Problem statement (with data from Phase 4 interviews)
  - Interaction model (from 3.1)
  - Architecture (from Phase 1 implementation)
  - Evaluation (from Phase 1 benchmark results)
  - Comparison with existing approaches
- [ ] Peer review with 2-3 distributed systems researchers
- [ ] Publish (arXiv or similar)

### Deliverables

- Interaction Model Specification (versioned document)
- Pattern Taxonomy Reference
- Governance Policy Language Spec
- Published Whitepaper

---

# Phase 2 — Fundraising Narrative (Fourth)

> **Goal:** Position ClawNet as "the control plane for AI agent ecosystems" with validated evidence.

### 2.1 Narrative Construction

- [ ] Craft core positioning: ClawNet = **the control plane for AI agent ecosystems**
- [ ] Build story arc:
  1. The problem (from Phase 4 market validation data)
  2. The insight (interaction observability, not agent intelligence)
  3. The product (from Phase 1 working MVP)
  4. The moat (from Phase 3 formal specification)
- [ ] Prepare one-liner, elevator pitch, and full narrative versions

### 2.2 Evidence Package

- [ ] Compile design partner testimonials / case studies
- [ ] Document quantitative results:
  - Debug time reduction (before/after ClawNet)
  - Cost savings from loop/fan-out detection
  - Failure attribution accuracy
- [ ] Prepare live demo script (trace a multi-agent failure end-to-end)

### 2.3 Market Sizing & Business Model

- [ ] Size the TAM/SAM/SOM for multi-agent observability tooling
- [ ] Define pricing model (usage-based? seat-based? hybrid?)
- [ ] Identify go-to-market strategy:
  - Open-source core + commercial cloud
  - Developer-first adoption (bottom-up)
  - Enterprise sales (top-down)

### 2.4 Fundraising Materials

- [ ] Pitch deck (12-15 slides)
- [ ] Financial model (projections, unit economics)
- [ ] Technical deep-dive appendix (for technical investors)
- [ ] Demo environment (hosted, one-click)

### 2.5 Investor Outreach

- [ ] Build target investor list (AI infra, developer tools, enterprise SaaS)
- [ ] Warm introductions through design partners and advisors
- [ ] Fundraising timeline and process management

### Deliverables

- Pitch deck
- Financial model
- Live demo environment
- Evidence package (testimonials + metrics)
- Target investor list

---

# Summary

```
WE ARE HERE
    |
    v
Phase 4: Market Validation .............. [ NOT STARTED ]
    |
    v
Phase 1: Interaction Trace MVP .......... [ IN PROGRESS — Skeleton Complete ]
    |                                       ~459 lines of Go
    |                                       Core: protocol, eventstore, task FSM,
    |                                              kernel, attribution
    |                                       Working demo: 3-agent failure scenario
    |                                       Next: Tracer SDK, Collector API, Postgres,
    |                                              pattern detection, replay
    v
Phase 3: Formal Specification ........... [ NOT STARTED ]
    |
    v
Phase 2: Fundraising Narrative .......... [ NOT STARTED ]
```

---

## Core Principle

ClawNet does not add complexity.

**It exposes the complexity that already exists.**
