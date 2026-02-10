1. ClawNet 是一個 multi-agent interaction kernel，不是 agent 框架。
2. 它把 agent 之間的互動轉換成 可觀測、可重播、可歸責的工程事件。
3. 當多 agent 系統失敗時，ClawNet 能精準指出 哪個 agent、在哪個互動、違反了什麼不變量。
4. 這讓 multi-agent 系統第一次能被 debug、replay、寫 postmortem、做 regression test。
5. ClawNet 的目標不是讓 agent 變聰明，而是讓 agent 互動變得可控、可負責、可演進。

# ClawNet (Engineering Skeleton)

This repository is a **Phase 1-first** engineering skeleton for ClawNet: an agent-to-agent communication kernel that makes interaction **observable**, **replayable**, and **attributable**.

## Quick start

```bash
go run ./cmd/clawnetd
```

You should see:
- a small 3-agent run (`Planner`, `Worker`, `Verifier`)
- an intentional failure (`INSUFFICIENT_EVIDENCE`)
- a generated **FailureReport** with a responsibility chain

## Repo layout

- `internal/protocol` — intents, messages
- `internal/eventstore` — append-only event log (in-memory v0)
- `internal/task` — task state machine
- `internal/kernel` — message routing + state transitions + event emission
- `internal/attribution` — failure classification & responsibility chain
- `examples/simple_task` — minimal runnable scenario

## Next milestones (post-skeleton)

- Replace in-memory store with Postgres
- Add strict replay (event-only) and soft replay (LLM) modes
- Add metrics extractor (Phase 2)
- Add reputation & teaming (Phase 3)

## Why ClawNet?
Multi-agent systems are increasingly used to solve complex tasks, but most existing frameworks fail at a fundamental engineering requirement: When the system fails, we cannot explain why in an actionable way.

Today, multi-agent failures are typically described as: 
**“the model hallucinated”** 
**“agents misunderstood each other”** 
**“the prompt needs tuning”**

These explanations are not acceptable in production engineering, because they provide:
- no reproducible root cause
- no regression test
- no concrete corrective action

The core problem is not intelligence — it is observability.
Multi-agent systems are distributed decision systems:
- state is fragmented across agents
- decisions are interdependent
- errors emerge late and indirectly

Yet most frameworks treat agent interaction as unstructured chat logs. This is equivalent to running a distributed system without tracing, replay, or postmortems.

## What ClawNet changes
ClawNet introduces an interaction kernel that makes agent interaction:
- Observable — interactions are typed intents, not raw text
- Replayable — every run is event-sourced and deterministic
- Attributable — failures produce responsibility chains

Instead of asking “what did the model say?”, ClawNet answers:

- which agent
- at which interaction
- violated which invariant
This allows multi-agent systems to be:
- debugged
- regression tested
- safely evolved

## Why this is not over-engineering
Every mature engineering domain went through the same transition:
- Domain	Before	After
- Web	ad-hoc requests	HTTP specification
- Microservices	logs only	tracing & spans
- Distributed systems	best effort	consensus & invariants
- Multi-agent	prompt chaining	interaction kernel

**ClawNet does not add complexity —
it exposes the complexity that already exists.**
