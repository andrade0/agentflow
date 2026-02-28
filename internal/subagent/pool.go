// Package subagent handles subagent spawning and pool management
package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentflow/agentflow/internal/agent"
	"github.com/agentflow/agentflow/internal/provider"
	"github.com/agentflow/agentflow/internal/skill"
	"github.com/agentflow/agentflow/pkg/types"
)

// Task represents a task to be executed by a subagent
type Task struct {
	ID          string
	Description string
	SkillName   string
	Message     string
	Metadata    map[string]string
}

// Result represents the result of a subagent task
type Result struct {
	TaskID    string
	AgentID   string
	Response  *types.CompletionResponse
	Error     error
	Duration  time.Duration
	StartedAt time.Time
}

// Pool manages a pool of subagents
type Pool struct {
	mu          sync.RWMutex
	provider    provider.Provider
	model       string
	skills      *skill.Loader
	maxAgents   int
	activeCount int
	results     map[string]*Result
	systemPrompt string
}

// PoolConfig holds pool configuration
type PoolConfig struct {
	Provider     provider.Provider
	Model        string
	Skills       *skill.Loader
	MaxAgents    int
	SystemPrompt string
}

// NewPool creates a new subagent pool
func NewPool(cfg PoolConfig) *Pool {
	if cfg.MaxAgents <= 0 {
		cfg.MaxAgents = 5
	}
	return &Pool{
		provider:     cfg.Provider,
		model:        cfg.Model,
		skills:       cfg.Skills,
		maxAgents:    cfg.MaxAgents,
		results:      make(map[string]*Result),
		systemPrompt: cfg.SystemPrompt,
	}
}

// Spawn creates a new subagent and executes a task
func (p *Pool) Spawn(ctx context.Context, task Task) (*Result, error) {
	p.mu.Lock()
	if p.activeCount >= p.maxAgents {
		p.mu.Unlock()
		return nil, fmt.Errorf("pool exhausted: max %d agents", p.maxAgents)
	}
	p.activeCount++
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.activeCount--
		p.mu.Unlock()
	}()

	// Create fresh agent for this task
	agentID := fmt.Sprintf("subagent-%s-%d", task.ID, time.Now().UnixNano())
	
	systemPrompt := p.systemPrompt
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("You are a focused subagent executing task: %s", task.Description)
	}

	a := agent.New(agent.Config{
		ID:           agentID,
		Provider:     p.provider,
		Model:        p.model,
		Skills:       p.skills,
		SystemPrompt: systemPrompt,
		Metadata:     task.Metadata,
	})

	startedAt := time.Now()
	
	var resp *types.CompletionResponse
	var err error

	if task.SkillName != "" {
		resp, err = a.RunWithSkill(ctx, task.SkillName, task.Message)
	} else {
		resp, err = a.Run(ctx, task.Message)
	}

	result := &Result{
		TaskID:    task.ID,
		AgentID:   agentID,
		Response:  resp,
		Error:     err,
		Duration:  time.Since(startedAt),
		StartedAt: startedAt,
	}

	// Store result
	p.mu.Lock()
	p.results[task.ID] = result
	p.mu.Unlock()

	return result, err
}

// SpawnAsync spawns a subagent asynchronously
func (p *Pool) SpawnAsync(ctx context.Context, task Task) <-chan *Result {
	ch := make(chan *Result, 1)
	go func() {
		result, _ := p.Spawn(ctx, task)
		ch <- result
		close(ch)
	}()
	return ch
}

// SpawnBatch spawns multiple subagents for parallel execution
func (p *Pool) SpawnBatch(ctx context.Context, tasks []Task) []*Result {
	var wg sync.WaitGroup
	results := make([]*Result, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			result, _ := p.Spawn(ctx, t)
			results[idx] = result
		}(i, task)
	}

	wg.Wait()
	return results
}

// GetResult retrieves a stored result by task ID
func (p *Pool) GetResult(taskID string) (*Result, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result, ok := p.results[taskID]
	return result, ok
}

// ActiveCount returns the number of active subagents
func (p *Pool) ActiveCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.activeCount
}

// ClearResults clears stored results
func (p *Pool) ClearResults() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.results = make(map[string]*Result)
}

// Stats returns pool statistics
type Stats struct {
	Active    int
	MaxAgents int
	Results   int
}

func (p *Pool) Stats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return Stats{
		Active:    p.activeCount,
		MaxAgents: p.maxAgents,
		Results:   len(p.results),
	}
}
