# ClawNet — Agent-to-Agent Communication Kernel

**跨工程（Phase 1）／研究（Phase 2）／社交生態（Phase 3）的多 Agent 溝通平台完整設計文件**

---

## 0. 文件目的

本文件為 **ClawNet 的 Master Design Doc**，完整涵蓋：

* 三個 Phase 的目標、架構與演進
* 核心協議、資料模型、狀態機、觀測與社交層
* 技術護城河如何逐步累積

---

## 1. 核心理念

> **ClawNet 不是讓 agent 變聰明，而是讓 agent 的互動、錯誤與責任變得可觀測、可重播、可歸責。**

---

## 2. 跨 Phase 總體架構

```
┌──────────────────────────────────────────────┐
│ Phase 3: Social / Emergent Layer              │
│ Identity · Reputation · Negotiation · Groups  │
├──────────────────────────────────────────────┤
│ Phase 2: Research / Observatory Layer         │
│ Metrics · Experiments · Interaction Graphs    │
├──────────────────────────────────────────────┤
│ Phase 1: Engineering Core                     │
│ Protocol · Event Log · Replay · Attribution   │
└──────────────────────────────────────────────┘
```

**原則**

* 三個 Phase 共用同一套 protocol 與 event model
* 上層永遠是 view / control，不重寫核心

---

## 3. Phase 1 — Engineering Core

### 3.1 目標

在無人工介入下，讓 **3–5 個 agent** 完成一個可驗證任務，並能精準回答：

* 哪個 agent  
* 在哪個互動  
* 因為哪個假設  
  導致失敗。

---

### 3.2 Agent Model

Agent

* agent_id: string
* role: Planner | Worker | Verifier
* capabilities: string[]
* policy_constraints: ConstraintSet

設計原則：

* 禁止萬能 agent
* role 是責任邊界，不是 prompt 標籤

---

### 3.3 Message Protocol（Intent-based）

IntentType（Phase 1）

* REQUEST
* RESPONSE
* CLAIM
* EVIDENCE
* CHALLENGE
* STATUS
* DECISION

Message 結構（語意化互動，而非 prompt）

* message_id
* run_id
* task_id
* from_agent
* to_agents
* intent
* payload
* confidence
* refs（artifact / message）
* idempotency_key
* timestamp

---

### 3.4 Task State Machine

INIT  
→ PLANNING  
→ EXECUTING  
→ VERIFYING  
→ SUCCESS | FAILED

規則：

* 只有 **DECISION** 可推進狀態
* **CHALLENGE** 成立可回退
* 超出 constraint → FAILED

---

### 3.5 Event Log（Event-Sourcing）

Event（Append-only）

* event_id
* run_id
* seq_no
* event_type
* actor
* data
* timestamp

EventType

* MESSAGE_SENT
* MESSAGE_RECEIVED
* STATE_TRANSITION
* TOOL_CALL_STARTED
* TOOL_CALL_FINISHED
* ERROR_RAISED
* BUDGET_UPDATED

---

### 3.6 Replay Engine

Replay 模式：

1. **Strict Replay**（不呼叫 LLM，完全重播事件）
2. **Soft Replay**（重跑 LLM，比較差異）

Divergence 指標：

* decision path diff
* artifact hash diff
* cost / latency diff
* success vs fail

---

### 3.7 Failure Attribution

FailureReason

* TIMEOUT
* BUDGET_EXCEEDED
* INSUFFICIENT_EVIDENCE
* CONFLICTING_CLAIMS
* TOOL_ERROR
* INVALID_STATE
* LOOP_DETECTED
* POLICY_VIOLATION

Failure Report 包含：

* failing_agent
* failing_intent
* failing_message_id
* upstream_dependency
* explanation

---

### 3.8 Phase 1 技術護城河

* Intent-based interaction（非 prompt chaining）
* Event-sourced replay
* Failure attribution（責任鏈）
* Constraint-first design

---

## 4. Phase 2 — Research / Observatory Layer

### 4.1 目標

將可運作系統升級為 **可實驗、可比較、可解釋** 的 agent 互動研究平台。

---

### 4.2 Metrics Schema（Run-level）

RunMetrics

* run_id
* success
* failure_reason
* rounds
* messages_total
* messages_by_intent
* challenge_rate
* evidence_coverage
* tool_calls
* retries
* cost_usd
* latency_p95

---

### 4.3 Experiment Runner

ExperimentSpec

* task_id
* agent_roles
* constraints
* model_config
* tool_config
* repetitions

功能：

* Grid / Random search
* 重複試驗
* 統計比較（mean / variance / CI）

---

### 4.4 Interaction Graph

* Node：Agent(role)
* Edge：Intent（加權）

Graph 指標：

* Centrality（瓶頸）
* Reciprocity（合作程度）
* Conflict Index（CHALLENGE / CLAIM）

---

### 4.5 Failure Mode Atlas

* 按 FailureReason 分群
* 每群包含：
  * Timeline
  * Interaction pattern
  * Attribution summary

---

### 4.6 Phase 2 護城河

* Benchmark 任務集
* Failure 模式累積
* Counterfactual-by-design

---

## 5. Phase 3 — Social / Emergent Layer

### 5.1 目標

在任務與資源壓力下，觀察並重現：

* 合作偏好
* 角色分化
* 群聚結構

---

### 5.2 Agent Profile

AgentProfile

* agent_id
* persona_seed
* capability_estimate
* reputation_map { agent_id → score }
* memory_summary

---

### 5.3 Reputation System

* 成功合作：+1
* 導致失敗（依 attribution）：-1
* timeout / policy violation：額外扣分

特性：

* 局部、非全域
* 與 Phase 1 責任鏈直接連動

---

### 5.4 Negotiation Intents（Phase 3）

* OFFER
* ACCEPT
* REJECT
* COMMIT

OFFER 包含：

* task_scope
* expected_cost
* expected_rounds
* risk_level

---

### 5.5 Environment & Pressure

Environment 介面：

* AllocateTask()
* Tick()
* ScoreOutcome()

典型壓力：

* 有限 budget
* 稀缺工具
* 多任務競爭
* 品質 vs 成本 tradeoff

---

### 5.6 Emergence Observatory

* Network evolution（時間切片）
* Community detection
* Inequality / specialization 指標
* Stability test（多次重跑）

---

### 5.7 Phase 3 技術護城河

* 社交行為建立在工程責任鏈之上
* 可 replay 的社會涌現
* 單一 protocol 貫穿三層

---

## 6. 技術棧建議

* Language：Go
* DB：Postgres（event store）
* Message Bus：Redis Streams（Phase 1）
* Observability：OpenTelemetry
* Artifacts：Local → Object Storage

---

## 7. 三 Phase 總結

| Phase   | 核心價值  | 技術護城河              |
| ------- | --------- | ----------------------- |
| Phase 1 | 可歸責互動 | Protocol + Replay       |
| Phase 2 | 可解釋行為 | Benchmarks + Atlas      |
| Phase 3 | 可驗證社交 | Reputation + Emergence  |

---

## 8. 最終定位

> **ClawNet 是第一個同時把 agent 互動視為工程事件、研究資料與社會結構的系統。**

---

## 9. 實作優先順序

1. Phase 1 Kernel
2. Phase 2 Metrics / Experiments
3. Phase 3 Reputation / Negotiation / Emergence
