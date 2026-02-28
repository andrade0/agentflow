package subagent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agentflow/agentflow/pkg/types"
)

// mockProvider implements provider.Provider for testing
type mockProvider struct {
	name     string
	response string
	delay    time.Duration
	err      error
	calls    int32
}

func (m *mockProvider) Name() string         { return m.name }
func (m *mockProvider) Models() []string     { return []string{"test-model"} }
func (m *mockProvider) SupportsModel(string) bool { return true }

func (m *mockProvider) Complete(ctx context.Context, req types.CompletionRequest) (*types.CompletionResponse, error) {
	atomic.AddInt32(&m.calls, 1)
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	return &types.CompletionResponse{
		Content:      m.response,
		Model:        req.Model,
		FinishReason: "stop",
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req types.CompletionRequest) (<-chan types.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

func TestNewPool(t *testing.T) {
	p := &mockProvider{name: "test"}
	pool := NewPool(PoolConfig{
		Provider:  p,
		Model:     "test-model",
		MaxAgents: 3,
	})

	if pool == nil {
		t.Fatal("expected non-nil pool")
	}

	stats := pool.Stats()
	if stats.MaxAgents != 3 {
		t.Errorf("MaxAgents = %d, want 3", stats.MaxAgents)
	}
	if stats.Active != 0 {
		t.Errorf("Active = %d, want 0", stats.Active)
	}
}

func TestPool_DefaultMaxAgents(t *testing.T) {
	p := &mockProvider{name: "test"}
	pool := NewPool(PoolConfig{
		Provider: p,
		Model:    "test",
	})

	stats := pool.Stats()
	if stats.MaxAgents != 5 {
		t.Errorf("default MaxAgents = %d, want 5", stats.MaxAgents)
	}
}

func TestPool_Spawn(t *testing.T) {
	p := &mockProvider{name: "test", response: "Task completed!"}
	pool := NewPool(PoolConfig{
		Provider:  p,
		Model:     "test-model",
		MaxAgents: 5,
	})

	task := Task{
		ID:          "task-1",
		Description: "Test task",
		Message:     "Do something",
	}

	result, err := pool.Spawn(context.Background(), task)
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}

	if result.TaskID != "task-1" {
		t.Errorf("TaskID = %q", result.TaskID)
	}
	if result.AgentID == "" {
		t.Error("expected non-empty AgentID")
	}
	if result.Response.Content != "Task completed!" {
		t.Errorf("Content = %q", result.Response.Content)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestPool_SpawnWithError(t *testing.T) {
	expectedErr := errors.New("provider error")
	p := &mockProvider{name: "test", err: expectedErr}
	pool := NewPool(PoolConfig{Provider: p, Model: "test"})

	task := Task{ID: "error-task", Message: "This will fail"}

	result, err := pool.Spawn(context.Background(), task)
	if err == nil {
		t.Fatal("expected error")
	}

	if result == nil {
		t.Fatal("expected result even on error")
	}
	if result.Error == nil {
		t.Error("expected result.Error")
	}
}

func TestPool_MaxAgentsLimit(t *testing.T) {
	p := &mockProvider{name: "test", response: "ok", delay: 100 * time.Millisecond}
	pool := NewPool(PoolConfig{
		Provider:  p,
		Model:     "test",
		MaxAgents: 2,
	})

	ctx := context.Background()
	
	// Start 2 tasks that will be slow
	done := make(chan struct{})
	go func() {
		pool.Spawn(ctx, Task{ID: "slow-1", Message: "slow"})
		done <- struct{}{}
	}()
	go func() {
		pool.Spawn(ctx, Task{ID: "slow-2", Message: "slow"})
		done <- struct{}{}
	}()

	// Give them time to start
	time.Sleep(20 * time.Millisecond)

	// Third task should fail immediately
	_, err := pool.Spawn(ctx, Task{ID: "overflow", Message: "overflow"})
	if err == nil {
		t.Error("expected pool exhausted error")
	}

	// Wait for first two to complete
	<-done
	<-done
}

func TestPool_SpawnAsync(t *testing.T) {
	p := &mockProvider{name: "test", response: "async result"}
	pool := NewPool(PoolConfig{Provider: p, Model: "test"})

	task := Task{ID: "async-1", Message: "async task"}
	ch := pool.SpawnAsync(context.Background(), task)

	result := <-ch
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Response.Content != "async result" {
		t.Errorf("Content = %q", result.Response.Content)
	}
}

func TestPool_SpawnBatch(t *testing.T) {
	p := &mockProvider{name: "test", response: "batch result"}
	pool := NewPool(PoolConfig{Provider: p, Model: "test", MaxAgents: 10})

	tasks := []Task{
		{ID: "batch-1", Message: "task 1"},
		{ID: "batch-2", Message: "task 2"},
		{ID: "batch-3", Message: "task 3"},
	}

	results := pool.SpawnBatch(context.Background(), tasks)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result == nil {
			t.Errorf("result[%d] is nil", i)
			continue
		}
		if result.TaskID != tasks[i].ID {
			t.Errorf("result[%d].TaskID = %q, want %q", i, result.TaskID, tasks[i].ID)
		}
	}

	if atomic.LoadInt32(&p.calls) != 3 {
		t.Errorf("expected 3 provider calls, got %d", p.calls)
	}
}

func TestPool_GetResult(t *testing.T) {
	p := &mockProvider{name: "test", response: "stored"}
	pool := NewPool(PoolConfig{Provider: p, Model: "test"})

	task := Task{ID: "store-1", Message: "store this"}
	pool.Spawn(context.Background(), task)

	result, ok := pool.GetResult("store-1")
	if !ok {
		t.Fatal("expected result to be stored")
	}
	if result.Response.Content != "stored" {
		t.Errorf("Content = %q", result.Response.Content)
	}

	// Non-existent task
	_, ok = pool.GetResult("nonexistent")
	if ok {
		t.Error("expected no result for nonexistent task")
	}
}

func TestPool_ClearResults(t *testing.T) {
	p := &mockProvider{name: "test", response: "ok"}
	pool := NewPool(PoolConfig{Provider: p, Model: "test"})

	pool.Spawn(context.Background(), Task{ID: "clear-1", Message: "a"})
	pool.Spawn(context.Background(), Task{ID: "clear-2", Message: "b"})

	stats := pool.Stats()
	if stats.Results != 2 {
		t.Errorf("Results = %d, want 2", stats.Results)
	}

	pool.ClearResults()

	stats = pool.Stats()
	if stats.Results != 0 {
		t.Errorf("Results after clear = %d, want 0", stats.Results)
	}
}

func TestPool_ActiveCount(t *testing.T) {
	p := &mockProvider{name: "test", response: "ok", delay: 50 * time.Millisecond}
	pool := NewPool(PoolConfig{Provider: p, Model: "test", MaxAgents: 5})

	// Initially zero
	if pool.ActiveCount() != 0 {
		t.Errorf("initial ActiveCount = %d", pool.ActiveCount())
	}

	// Start async task
	pool.SpawnAsync(context.Background(), Task{ID: "active-1", Message: "slow"})
	time.Sleep(10 * time.Millisecond)

	if pool.ActiveCount() != 1 {
		t.Errorf("ActiveCount during task = %d", pool.ActiveCount())
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	if pool.ActiveCount() != 0 {
		t.Errorf("ActiveCount after task = %d", pool.ActiveCount())
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	p := &mockProvider{name: "test", response: "ok", delay: 1 * time.Second}
	pool := NewPool(PoolConfig{Provider: p, Model: "test"})

	ctx, cancel := context.WithCancel(context.Background())
	
	done := make(chan struct{})
	go func() {
		pool.Spawn(ctx, Task{ID: "cancel-1", Message: "long task"})
		done <- struct{}{}
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good, task was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("task should have been cancelled")
	}
}
