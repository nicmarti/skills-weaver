package agent

import (
	"fmt"
	"time"
)

// AgentError represents an error that occurred during agent operations.
type AgentError struct {
	AgentName string
	Operation string
	Err       error
}

func (e *AgentError) Error() string {
	return fmt.Sprintf("agent %s: %s failed: %v", e.AgentName, e.Operation, e.Err)
}

func (e *AgentError) Unwrap() error {
	return e.Err
}

// ErrAgentNotFound indicates that the requested agent does not exist.
type ErrAgentNotFound struct {
	AgentName      string
	AvailableAgents []string
}

func (e *ErrAgentNotFound) Error() string {
	return fmt.Sprintf("agent not found: %s (available: %v)", e.AgentName, e.AvailableAgents)
}

// ErrRecursionLimit indicates that the agent invocation depth limit was exceeded.
type ErrRecursionLimit struct {
	AgentName    string
	CurrentDepth int
	MaxDepth     int
	CallChain    []string
}

func (e *ErrRecursionLimit) Error() string {
	return fmt.Sprintf("recursion limit exceeded: agent %s at depth %d (max: %d), call chain: %v",
		e.AgentName, e.CurrentDepth, e.MaxDepth, e.CallChain)
}

// ErrAgentTimeout indicates that an agent invocation timed out.
type ErrAgentTimeout struct {
	AgentName string
	Timeout   time.Duration
}

func (e *ErrAgentTimeout) Error() string {
	return fmt.Sprintf("agent %s timed out after %v", e.AgentName, e.Timeout)
}

// SkillError represents an error that occurred during skill operations.
type SkillError struct {
	SkillName string
	Operation string
	Err       error
}

func (e *SkillError) Error() string {
	return fmt.Sprintf("skill %s: %s failed: %v", e.SkillName, e.Operation, e.Err)
}

func (e *SkillError) Unwrap() error {
	return e.Err
}

// ErrSkillNotFound indicates that the requested skill does not exist.
type ErrSkillNotFound struct {
	SkillName       string
	AvailableSkills []string
}

func (e *ErrSkillNotFound) Error() string {
	return fmt.Sprintf("skill not found: %s (available: %v)", e.SkillName, e.AvailableSkills)
}

// ErrSkillExecution indicates that a skill execution failed.
type ErrSkillExecution struct {
	SkillName  string
	Command    string
	ExitCode   int
	Stdout     string
	Stderr     string
	Err        error
}

func (e *ErrSkillExecution) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("skill %s execution failed: %v (command: %s)", e.SkillName, e.Err, e.Command)
	}
	return fmt.Sprintf("skill %s execution failed with exit code %d (command: %s, stderr: %s)",
		e.SkillName, e.ExitCode, e.Command, e.Stderr)
}

func (e *ErrSkillExecution) Unwrap() error {
	return e.Err
}

// ErrInvalidPersona indicates that a persona file is malformed.
type ErrInvalidPersona struct {
	PersonaName string
	Reason      string
}

func (e *ErrInvalidPersona) Error() string {
	return fmt.Sprintf("invalid persona %s: %s", e.PersonaName, e.Reason)
}

// ErrStatePersistence indicates that agent state could not be saved or loaded.
type ErrStatePersistence struct {
	Operation string // "save" or "load"
	FilePath  string
	Err       error
}

func (e *ErrStatePersistence) Error() string {
	return fmt.Sprintf("state persistence %s failed for %s: %v", e.Operation, e.FilePath, e.Err)
}

func (e *ErrStatePersistence) Unwrap() error {
	return e.Err
}
