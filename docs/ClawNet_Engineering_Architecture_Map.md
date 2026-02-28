# ClawNet Engineering Map (Phase 1 v0)

> Last updated: 2026-02-28 21:49 (UTC+8)

> **一份「最直覺」的架構圖 + 開發流程 + function 導覽 + protocol 定義 + 測試細節**
> 對應 repo：`clawnet_skeleton.zip`（Go）

---

## 0. 這個工具的用途（一句話）
**ClawNet 讓多 agent 系統的互動變成可觀測、可重播、可歸責的工程事件**：  
當 run 失敗時，你能回答：**誰（agent）、在哪個互動（intent/message）、為什麼（evidence/constraint/state）**。

---

## 1. 系統心智模型（Architecture Map）

### 1.1 三個角色 + 一個核心內核（Kernel）
```
Planner Agent      Worker Agent        Verifier Agent
    |                  |                   |
    | DECISION          | CLAIM             | CHALLENGE
    v                  v                   v
+------------------------------------------------------+
|                 Communication Kernel                 |
|   - validates state transitions (Task FSM)           |
|   - appends events (EventStore)                      |
|   - enforces constraints (rounds/messages/budget)    |
|   - produces FailureReport (Attribution v0)          |
+------------------------------------------------------+
                     |
                     v
                +----------+
                | EventLog  |
                | (RunID)   |
                +----------+
```

### 1.2 Phase 1 的「資料流」：Message → Event → Report
```
protocol.Message (intent, refs, payload)
          |
          v
kernel.EmitMessage(m)  --append-->  eventstore.Event (append-only)
          |
          +--(if DECISION)--> task.Apply(intent) --> EventStateTransition
          |
          +--(if constraint violated)--> ERROR event
          |
          v
kernel.Finalize...() -> attribution.FailureReport
```

> **設計理念**：Agent 只能「提出意圖」，Kernel 才能「改變世界狀態」並寫入 truth source（EventLog）。

---

## 2. Repo 結構（檔案 = 責任邊界）
```
clawnet/
├── cmd/clawnetd/main.go                    # 入口：跑 example 並輸出 FailureReport
├── internal/protocol/intent.go             # Intent enum（typed interaction）
├── internal/protocol/message.go            # Message schema（run/task/refs/...）
├── internal/eventstore/event.go            # Event schema（append-only）
├── internal/eventstore/store.go            # InMemoryStore（v0）
├── internal/task/state.go                  # Task FSM states
├── internal/task/transition.go             # Allowed transitions + Apply()
├── internal/constraints/constraints.go     # MaxRounds/MaxMessages/Budget/Deadline
├── internal/kernel/kernel.go               # Kernel：message→event→state→constraints
├── internal/attribution/failure.go         # FailureReason + FailureReport
└── examples/simple_task/simple_task.go     # 最小失敗案例（最重要的驗證點）
```

---

## 3. Protocol 定義（Phase 1）

### 3.1 IntentType（語意化互動，而非 prompt）
檔案：`internal/protocol/intent.go`
- `REQUEST`
- `RESPONSE`
- `CLAIM`
- `EVIDENCE`
- `CHALLENGE`
- `STATUS`
- `DECISION`

### 3.2 Message（最小但不會後悔的 schema）
檔案：`internal/protocol/message.go`

必備欄位（Phase 1 的可觀測性核心）：
- `MessageID`：每次互動的唯一 ID（可追責）
- `RunID` / `TaskID`：把訊息歸到一個 run / task（可 replay）
- `FromAgent` / `ToAgents`：責任與路由
- `Intent`：互動語意（可計算、可統計、可約束）
- `Payload`：內容（可變動，不影響系統層）
- `Refs`：**責任鏈依賴**（Phase 1 的 attribution 會用到）
- `Timestamp`：時序分析

> **要點**：ClawNet 的護城河是「intent + refs」，不是 payload。

---

## 4. Event Sourcing（為什麼要有 EventLog）

### 4.1 Event schema
檔案：`internal/eventstore/event.go`

EventType（v0）
- `MESSAGE_SENT`
- `STATE_TRANSITION`
- `ERROR_RAISED`
- `BUDGET_UPDATED`（預留）

Event 欄位
- `RunID`：run 索引
- `SeqNo`：序列（重播順序）
- `Actor`：誰造成這個 event
- `Data`：可序列化內容（message / state string / error string）

### 4.2 Store (v0 InMemory)
檔案：`internal/eventstore/store.go`
- `Append(runID, e)`：append-only + per-run seq increment
- `List(runID)`：讀出整串 event（可做 replay）

> **設計理念**：不要靠 printf log；log 本身就是資料，才能做 replay/analysis。

---

## 5. Task FSM（防 agent 亂講、亂結案）

檔案：`internal/task/state.go`, `internal/task/transition.go`

### 5.1 狀態
- `INIT`
- `PLANNING`
- `EXECUTING`
- `VERIFYING`
- `SUCCESS`
- `FAILED`（v0 暫時由 report 表達，未硬落 state）

### 5.2 Transition rules（v0）
檔案：`transition.go` 的 `allowed` map 定義：
- `INIT` 只接受 `DECISION` → `PLANNING`
- `PLANNING` 只接受 `DECISION` → `EXECUTING`
- `EXECUTING` 接受 `DECISION` → `VERIFYING`
- `VERIFYING` 接受 `DECISION` → `SUCCESS`
- `CHALLENGE` 目前不強制 rollback（**待 v1 強化**）

**關鍵函式**
- `(*task.Task).Apply(intent)`：檢查 intent 是否允許推進狀態，不允許則回 error

> **設計理念**：只有 `DECISION` 能改變世界狀態；否則你永遠在聊天。

---

## 6. Kernel（Phase 1 的「法律」）

檔案：`internal/kernel/kernel.go`

### 6.1 Kernel 的責任
- **寫入事件**：每個 message 進來都要 append event
- **推進狀態**：只在 `DECISION` 時呼叫 `task.Apply()`
- **記錄 StateTransition event**
- **檢查 constraints**
- **產出 failure report（v0）**

### 6.2 核心 function 導覽

#### `kernel.New(runID, taskID, store, constraints)`
建立一次 run 的核心內核：
- 綁定 event store
- 建立 task state machine
- 初始化 runID/taskID

#### `(*Kernel).EmitMessage(m protocol.Message) error`
**Phase 1 的關鍵入口**：把 agent 的互動「變成工程事件」

流程：
1. Append `MESSAGE_SENT` event（Data = message）
2. 若 `m.Intent == CLAIM`：記下 lastClaim（v0 attribution 用）
3. 若 `m.Intent == DECISION`：
   - `task.Apply(DECISION)` → 更新 state + round
   - Append `STATE_TRANSITION` event
4. `checkConstraints()`：MaxRounds / MaxMessages 等

#### `(*Kernel).FinalizeInsufficientEvidence(challenger, challengeMsgID)`
v0 attribution：把「最後一個 CLAIM」當作 failing point，產出：
- `Reason = INSUFFICIENT_EVIDENCE`
- `UpstreamChain = [claimID, challengeID]`
- `FailingAgent = lastClaimAgent`

> **設計理念**：先固定 FailureReport 的 interface，再把推導變聰明（v1 做 dependency graph trace）。

---

## 7. Constraints（把現實世界變成系統的一部分）

檔案：`internal/constraints/constraints.go`

v0 提供：
- `MaxRounds`
- `MaxMessages`
- `MaxBudgetUSDCents`（預留）
- `Deadline`（預留）

**關鍵函式**
- `constraints.Default()`：給範例跑起來的 default 限制

> **設計理念**：沒有 constraints，就沒有「在壓力下的真實互動」；Phase 2/3 也無法成立。

---

## 8. Attribution（失敗要有形狀）

檔案：`internal/attribution/failure.go`

### 8.1 FailureReason（v0）
- `INSUFFICIENT_EVIDENCE`
- `CONFLICTING_CLAIMS`
- `TIMEOUT`
- `BUDGET_EXCEEDED`
- `INVALID_STATE`
- `LOOP_DETECTED`
- `POLICY_VIOLATION`
- ...（預留）

### 8.2 FailureReport（Phase 1 output）
- `FailingAgent`
- `FailingIntent`
- `MessageID`
- `UpstreamChain`
- `Reason`
- `Explanation`

> **你拿這個 report 去跟別人說 ClawNet 是什麼，比講 100 句願景都有效。**

---

## 9. 開發流程（最直覺的「你要怎麼做」）

### 9.1 你現在（v0）能做的事
1. 跑起來：`go run ./cmd/clawnetd`
2. 看到 FailureReport
3. 確認：失敗原因 + failing agent 是否符合預期

### 9.2 v0 → v1（最建議的升級順序）
**Goal：把 heuristic attribution 變成真責任鏈**

1) **建立 dependency graph**
- 來源：`protocol.Message.Refs`
- 定義：message 是 node，refs 是 directed edges

2) **Attribution 追溯**
- 從 failing message 出發沿 refs 回溯
- 找到最早的無效來源（例如 evidence 缺失）

3) **CHALLENGE rollback**
- 在 `VERIFYING` 收到 challenge → 回到 `EXECUTING`
- 強迫 planner 再做 DECISION 才能回 VERIFYING

4) **EventStore 落到 Postgres**
- schema：`events(run_id, seq_no, event_id, type, actor, data, ts)`
- index：`(run_id, seq_no)`

---

## 10. 測試細節（怎麼測才算「工程級」）

> Phase 1 的測試重點不是 LLM 正確率，而是：**互動結構正確、事件可重播、失敗可歸責**。

### 10.1 建議測試分類

#### A. Protocol tests
- Intent enum 穩定性（新增不破壞舊值）
- Message 必填欄位檢查（run_id/task_id/from_agent/intent）

#### B. FSM tests (`internal/task`)
- `Apply(DECISION)` 是否依 state 正確轉移
- 非允許 intent 是否回 error
- round 是否正確遞增

#### C. EventStore tests (`internal/eventstore`)
- Append 是否 seq_no 遞增
- List 是否按 seq_no 返回
- 不同 run_id 不互相污染

#### D. Kernel tests (`internal/kernel`)
- EmitMessage 會 append MESSAGE_SENT event
- DECISION 會 append STATE_TRANSITION event
- invalid transition 會 append ERROR_RAISED（並返回 error）
- constraints 超限時必須 fail（返回 error + event）

#### E. Attribution tests (`internal/attribution`)
- v0：當 CLAIM 沒 refs，Verifier challenge → Reason=INSUFFICIENT_EVIDENCE
- v1（未做）：dependency graph trace 應該能找 root cause

### 10.2 最小測試案例（推薦你第一個寫）
檔案：`examples/simple_task/simple_task_test.go`（建議新增）

測試內容：
- Run() 回傳的 FailureReport
- `Reason == INSUFFICIENT_EVIDENCE`
- `FailingAgent == "worker"`
- `FailingIntent == CLAIM`
- `UpstreamChain` 包含 claimID 與 challengeID（長度 2）

---

## 11. 你要怎麼向外界解釋 ClawNet（最直覺的一段）
ClawNet 把 multi-agent 的互動定義成 **intent-based protocol**，並把每次互動寫入 **append-only event log**。  
因此每個 run 都可以被 replay，且失敗可以產出 **FailureReport**（誰、在哪個互動、為什麼）。  
這讓 multi-agent 從「prompt demo」變成「能 debug、能 postmortem 的工程系統」。

---

## 12. Appendix：關鍵函式索引（快速定位）
- Entry：`cmd/clawnetd/main.go`
- Run scenario：`examples/simple_task/simple_task.go::Run()`
- Kernel ingress：`internal/kernel/kernel.go::EmitMessage()`
- FSM：`internal/task/transition.go::Apply()`
- Event store：`internal/eventstore/store.go::Append(), List()`
- Attribution：`internal/attribution/failure.go::FailureReport`

---

## 13. 目前進度與下一步

### 已完成（~459 行 Go）

| 元件 | 狀態 | 說明 |
|------|------|------|
| `protocol/intent.go` | Done | 7 IntentTypes |
| `protocol/message.go` | Done | Message schema with refs |
| `eventstore/event.go` | Done | Event struct & 5 EventTypes |
| `eventstore/store.go` | Done | InMemoryStore (thread-safe) |
| `task/state.go` | Done | 6 states FSM |
| `task/transition.go` | Done | FSM validation + Apply() |
| `constraints/constraints.go` | Done | ConstraintSet defaults |
| `kernel/kernel.go` | Partial | 僅處理 DECISION intent |
| `attribution/failure.go` | Partial | 僅實現 INSUFFICIENT_EVIDENCE |
| `examples/simple_task` | Done | 3-agent 失敗場景 demo |

### v0 → v1 升級路線

1. **完善 Kernel** — 處理所有 intent types，啟用 BUDGET_UPDATED
2. **Tracer SDK** — 實現 StartTrace / StartSpan / Exporter
3. **Collector API** — POST /ingest + GET /traces/{trace_id}
4. **Postgres** — 替換 InMemoryStore
5. **Pattern Detection** — loop / fan-out / cost attribution
6. **Replay** — strict + soft 模式
7. **測試** — unit + integration + failure scenario tests

> 完整路線圖請參閱 [README.md](../README.md)

---
