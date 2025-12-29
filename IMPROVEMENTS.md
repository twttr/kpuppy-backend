# Code Review & Improvements Plan

## Current Stats
- **Lines of Code**: ~3,600
- **Tests**: 82 passing
- **Coverage**:
  - middleware: 91.2%
  - usecase: 79.0% ✓
  - domain: 66.7%
  - repository/sqlite: 66.1%
  - websocket: 57.8% ✓
  - handlers: 31.4% ✓

---

## Issues Found

### 1. [BUG] WebSocket broadcasts broken for replies/updates/deletes
**File**: `internal/handler/comment_handler.go:176-178`
**Problem**: `getContentByCommentID` always returns `nil`, so broadcasts never fire for replies, updates, or deletes.
**Impact**: Real-time updates don't work except for new top-level comments.

### 2. [SECURITY] No HTTP server timeouts
**File**: `cmd/server/main.go`
**Problem**: Echo server started without read/write timeouts.
**Impact**: Vulnerable to slowloris attacks.

### 3. [BEST PRACTICE] Error handling not idiomatic Go 1.13+
**File**: `internal/handler/comment_handler.go:180-201`
**Problem**: Using `switch err` with sentinel errors instead of `errors.Is()`.
**Impact**: Won't work correctly with wrapped errors.

### 4. [SIMPLIFY] Duplicate pagination validation
**Files**: `internal/handler/comment_handler.go` and `internal/usecase/comment_service.go`
**Problem**: Same validation logic in both handler and service.
**Impact**: Code duplication, potential for inconsistency.

### 5. [QUALITY] Low test coverage
**Problem**: handlers (2.8%), usecase (0%), websocket (0%)
**Impact**: Bugs may go undetected, harder to refactor safely.

### 6. [NICE-TO-HAVE] No structured logging
**File**: `cmd/server/main.go`
**Problem**: Using `fmt.Fprintf` for error output.
**Impact**: Harder to parse logs in production.

---

## Fixes Plan

### Priority 1: Bug Fix
- [x] Fix `getContentByCommentID` to actually fetch content
- [x] Add `GetContentByID` method to CommentService

### Priority 2: Security
- [x] Add HTTP server timeouts (ReadTimeout: 15s, WriteTimeout: 15s, IdleTimeout: 60s)

### Priority 3: Best Practices
- [x] Replace `switch err` with `errors.Is()` pattern (comment_handler, admin_handler, comment_service)
- [x] Remove duplicate validation (keep in service layer only)

### Priority 4: Quality
- [x] Add usecase tests (mock repositories) - 23 tests, 79% coverage
- [x] Add handler integration tests - 11 tests, 31.4% coverage
- [x] Add websocket tests - 7 tests, 57.8% coverage

### Priority 5: Nice-to-Have
- [x] Add `log/slog` for structured logging (JSON output)
- [x] Add request ID middleware for tracing
- [x] Integrate Sentry SDK (compatible with self-hosted Bugsink)

---

## Implementation Order

1. ~~Fix WebSocket bug (breaks real-time features)~~ ✓
2. ~~Add HTTP timeouts (security)~~ ✓
3. ~~Refactor error handling (best practice)~~ ✓
4. ~~Remove duplicate validation (simplify)~~ ✓
5. ~~Improve test coverage (quality)~~ ✓
6. ~~Add structured logging + Sentry~~ ✓

**All improvements complete!**
