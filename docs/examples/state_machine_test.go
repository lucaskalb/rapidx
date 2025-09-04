//go:build examples
// +build examples

package examples

import (
	"errors"
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

// BankAccount represents a simple bank account state machine.
type BankAccount struct {
	Balance int
	Closed  bool
}

// BankCommand represents commands that can be executed on a bank account.
type BankCommand struct {
	Type   string // "deposit", "withdraw", "close"
	Amount int
}

// TestBankAccount demonstrates state machine testing with a bank account.
func TestBankAccount(t *testing.T) {
	sm := prop.StateMachine[BankAccount, BankCommand]{
		InitialState: BankAccount{Balance: 0, Closed: false},
		Commands: []prop.Command[BankAccount, BankCommand]{
			{
				Name: "deposit",
				Generator: gen.Map(gen.IntRange(1, 1000), func(amount int) BankCommand {
					return BankCommand{Type: "deposit", Amount: amount}
				}),
				Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
					if state.Closed {
						return state, errors.New("account is closed")
					}
					return BankAccount{Balance: state.Balance + cmd.Amount, Closed: state.Closed}, nil
				},
				Precondition: func(state BankAccount, cmd BankCommand) bool {
					return !state.Closed
				},
				Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
					// After a deposit, balance should increase by the amount
					return to.Balance == from.Balance+cmd.Amount
				},
			},
			{
				Name: "withdraw",
				Generator: gen.Map(gen.IntRange(1, 1000), func(amount int) BankCommand {
					return BankCommand{Type: "withdraw", Amount: amount}
				}),
				Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
					if state.Closed {
						return state, errors.New("account is closed")
					}
					if state.Balance < cmd.Amount {
						return state, errors.New("insufficient funds")
					}
					return BankAccount{Balance: state.Balance - cmd.Amount, Closed: state.Closed}, nil
				},
				Precondition: func(state BankAccount, cmd BankCommand) bool {
					return !state.Closed && state.Balance >= cmd.Amount
				},
				Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
					// After a withdrawal, balance should decrease by the amount
					return to.Balance == from.Balance-cmd.Amount
				},
			},
			{
				Name:      "close",
				Generator: gen.Const(BankCommand{Type: "close", Amount: 0}),
				Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
					return BankAccount{Balance: state.Balance, Closed: true}, nil
				},
				Precondition: func(state BankAccount, cmd BankCommand) bool {
					return !state.Closed
				},
				Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
					// After closing, account should be closed but balance unchanged
					return to.Closed && to.Balance == from.Balance
				},
			},
		},
	}

	cfg := prop.Config{
		Seed:        12345,
		Examples:    10,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	prop.TestStateMachine(t, sm, cfg)
}

// Counter represents a simple counter state machine.
type Counter struct {
	Value int
}

// CounterCommand represents commands that can be executed on a counter.
type CounterCommand struct {
	Type  string // "increment", "decrement", "reset"
	Delta int
}

// TestCounter demonstrates state machine testing with a counter.
func TestCounter(t *testing.T) {
	sm := prop.StateMachine[Counter, CounterCommand]{
		InitialState: Counter{Value: 0},
		Commands: []prop.Command[Counter, CounterCommand]{
			{
				Name: "increment",
				Generator: gen.Map(gen.IntRange(1, 10), func(delta int) CounterCommand {
					return CounterCommand{Type: "increment", Delta: delta}
				}),
				Execute: func(state Counter, cmd CounterCommand) (Counter, error) {
					return Counter{Value: state.Value + cmd.Delta}, nil
				},
				Postcondition: func(from Counter, cmd CounterCommand, to Counter) bool {
					return to.Value == from.Value+cmd.Delta
				},
			},
			{
				Name: "decrement",
				Generator: gen.Map(gen.IntRange(1, 10), func(delta int) CounterCommand {
					return CounterCommand{Type: "decrement", Delta: delta}
				}),
				Execute: func(state Counter, cmd CounterCommand) (Counter, error) {
					return Counter{Value: state.Value - cmd.Delta}, nil
				},
				Postcondition: func(from Counter, cmd CounterCommand, to Counter) bool {
					return to.Value == from.Value-cmd.Delta
				},
			},
			{
				Name:      "reset",
				Generator: gen.Const(CounterCommand{Type: "reset", Delta: 0}),
				Execute: func(state Counter, cmd CounterCommand) (Counter, error) {
					return Counter{Value: 0}, nil
				},
				Postcondition: func(from Counter, cmd CounterCommand, to Counter) bool {
					return to.Value == 0
				},
			},
		},
	}

	cfg := prop.Config{
		Seed:        12345,
		Examples:    10,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	prop.TestStateMachine(t, sm, cfg)
}

// Cache represents a simple cache state machine.
type Cache struct {
	Data    map[string]string
	Size    int
	MaxSize int
}

// CacheCommand represents commands that can be executed on a cache.
type CacheCommand struct {
	Type  string // "get", "set", "delete", "clear"
	Key   string
	Value string
}

// TestCache demonstrates state machine testing with a cache.
func TestCache(t *testing.T) {
	sm := prop.StateMachine[Cache, CacheCommand]{
		InitialState: Cache{Data: make(map[string]string), Size: 0, MaxSize: 100},
		Commands: []prop.Command[Cache, CacheCommand]{
			{
				Name: "set",
				Generator: gen.Bind(
					gen.StringAlphaNum(gen.Size{Min: 1, Max: 10}),
					func(key string) gen.Generator[CacheCommand] {
						return gen.Map(
							gen.StringAlphaNum(gen.Size{Min: 1, Max: 20}),
							func(value string) CacheCommand {
								return CacheCommand{Type: "set", Key: key, Value: value}
							},
						)
					},
				),
				Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
					newState := state
					if newState.Data == nil {
						newState.Data = make(map[string]string)
					}
					// Create a new map to avoid mutation
					newData := make(map[string]string)
					for k, v := range state.Data {
						newData[k] = v
					}

					_, exists := newData[cmd.Key]
					newData[cmd.Key] = cmd.Value

					newState.Data = newData
					if !exists {
						newState.Size++
					}
					return newState, nil
				},
				Precondition: func(state Cache, cmd CacheCommand) bool {
					return state.Size < state.MaxSize || state.Data[cmd.Key] != ""
				},
				Postcondition: func(from Cache, cmd CacheCommand, to Cache) bool {
					// After setting, the key should have the value
					return to.Data[cmd.Key] == cmd.Value
				},
			},
			{
				Name: "get",
				Generator: gen.Map(gen.StringAlphaNum(gen.Size{Min: 1, Max: 10}), func(key string) CacheCommand {
					return CacheCommand{Type: "get", Key: key, Value: ""}
				}),
				Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
					// Get doesn't change state, just returns the current state
					return state, nil
				},
				Postcondition: func(from Cache, cmd CacheCommand, to Cache) bool {
					// Get should not change the state
					return to.Size == from.Size && len(to.Data) == len(from.Data)
				},
			},
			{
				Name: "delete",
				Generator: gen.Map(gen.StringAlphaNum(gen.Size{Min: 1, Max: 10}), func(key string) CacheCommand {
					return CacheCommand{Type: "delete", Key: key, Value: ""}
				}),
				Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
					newState := state
					if newState.Data == nil {
						newState.Data = make(map[string]string)
					}
					// Create a new map to avoid mutation
					newData := make(map[string]string)
					for k, v := range state.Data {
						newData[k] = v
					}

					_, exists := newData[cmd.Key]
					if exists {
						delete(newData, cmd.Key)
						newState.Size--
					}

					newState.Data = newData
					return newState, nil
				},
				Postcondition: func(from Cache, cmd CacheCommand, to Cache) bool {
					// After deletion, the key should not exist
					_, exists := to.Data[cmd.Key]
					return !exists
				},
			},
			{
				Name:      "clear",
				Generator: gen.Const(CacheCommand{Type: "clear", Key: "", Value: ""}),
				Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
					return Cache{Data: make(map[string]string), Size: 0, MaxSize: state.MaxSize}, nil
				},
				Postcondition: func(from Cache, cmd CacheCommand, to Cache) bool {
					// After clearing, size should be 0 and data should be empty
					return to.Size == 0 && len(to.Data) == 0
				},
			},
		},
	}

	cfg := prop.Config{
		Seed:        12345,
		Examples:    10,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	prop.TestStateMachine(t, sm, cfg)
}
