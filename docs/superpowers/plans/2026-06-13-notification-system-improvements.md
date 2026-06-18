# Notification System Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply code quality, performance, and simplicity improvements to the Go backend's follow/like/notification use cases and background worker.

**Architecture:** Reuse fetched domain objects to reduce redundant repository calls in usecases, implement defensive nil checks, and simplify concurrent workers using clean channel-range loops.

**Tech Stack:** Go 1.26+, GoMock, Testify.

---

### Task 1: Verify FollowUseCase Nil Pointer Check
**Files:**
- Modify: `internal/core/usecase/follow_usecase.go`

- [ ] **Step 1: Check existing defensive checks**
  Ensure the following nil checks are present in `Follow`:
  ```go
  if follower == nil || followed == nil {
      return nil, fmt.Errorf("get user: %w", domain.ErrUserNotFound)
  }
  ```
- [ ] **Step 2: Run follow usecase tests**
  Run: `go test -v -run TestFollowUseCase ./internal/core/usecase`
  Expected: PASS

---

### Task 2: Redundant DB Lookup in `PostUseCase.LikePost`
**Files:**
- Modify: `internal/core/usecase/post_usecase.go`
- Modify: `internal/core/usecase/post_usecase_test.go`

- [ ] **Step 1: Refactor `LikePost` in `internal/core/usecase/post_usecase.go`**
  Modify `LikePost` to:
  - Assign the result of the first `uc.postRepo.GetByIDOnly` to `post`.
  - Reuse `post` throughout the method instead of fetching it again.
  - When returning the updated like count on success, return `post.LikeCount + 1` instead of querying the post again.
  - Remove subsequent `GetByIDOnly` calls in `LikePost`.

- [ ] **Step 2: Update tests in `internal/core/usecase/post_usecase_test.go`**
  Modify `TestPostUseCase_LikePost` to match the reduced `GetByIDOnly` calls (only 1 expectation instead of 2).

- [ ] **Step 3: Run post usecase tests**
  Run: `go test -v -run TestPostUseCase ./internal/core/usecase`
  Expected: PASS

---

### Task 3: Simplify BackgroundWorker Loop using `range w.ch`
**Files:**
- Modify: `internal/adapters/notification/worker.go`
- Modify: `internal/adapters/notification/broker_worker_test.go`

- [ ] **Step 1: Refactor worker loop in `internal/adapters/notification/worker.go`**
  - Simplify BackgroundWorker struct to remove `quit` and `stopOnce` fields.
  - Modify `Start` to launch the range loop on `w.ch`.
  - Modify `Stop` to wait on `w.wg`.

- [ ] **Step 2: Refactor `TestBrokerAndWorker` in `internal/adapters/notification/broker_worker_test.go`**
  - Change the order in the test so that `broker.Close()` is called BEFORE `worker.Stop(ctx)`.

- [ ] **Step 3: Run notification adapter tests**
  Run: `go test -v ./internal/adapters/notification/...`
  Expected: PASS

---

### Task 4: Verify and Commit
- [ ] **Step 1: Run all tests in the workspace**
  Run: `go test ./...`
  Expected: PASS
- [ ] **Step 2: Commit changes**
  Run: `git commit -a -m "refactor: simplify worker channel loop, optimize LikePost queries, and add defensive checks in Follow"`
  Expected: Clean status, successful commit
