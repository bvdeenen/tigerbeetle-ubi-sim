/*
Package bank handles account lifecycle against TigerBeetle. Its Bootstrap function
ensures the central bank account and all agent accounts exist before the simulation
starts, creating any that are missing. It validates that existing accounts have the
expected ledger to catch stale data from previous runs early.
*/
package bank

import (
	"fmt"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

const (
	Ledger          uint32 = 1
	CodeCentralBank uint16 = 1
	CodeAgent       uint16 = 2
)

// Bootstrap ensures the central bank account and all agent accounts exist in TigerBeetle.
// It is idempotent and safe to call concurrently from multiple app instances.
func Bootstrap(client tb.Client, centralBankID types.Uint128, agentIDs []types.Uint128) error {
	allIDs := make([]types.Uint128, 0, 1+len(agentIDs))
	allIDs = append(allIDs, centralBankID)
	allIDs = append(allIDs, agentIDs...)

	existing, err := client.LookupAccounts(allIDs)
	if err != nil {
		return fmt.Errorf("lookup accounts: %w", err)
	}

	existingSet := make(map[types.Uint128]bool, len(existing))
	for _, a := range existing {
		if a.Ledger != Ledger {
			return fmt.Errorf("account %v already exists with ledger %d (expected %d) — "+
				"wipe the TigerBeetle data file or use a different --id-offset", a.ID, a.Ledger, Ledger)
		}
		existingSet[a.ID] = true
	}

	var toCreate []types.Account

	if !existingSet[centralBankID] {
		toCreate = append(toCreate, types.Account{
			ID:     centralBankID,
			Ledger: Ledger,
			Code:   CodeCentralBank,
			Flags:  types.AccountFlags{}.ToUint16(),
		})
	}

	for _, id := range agentIDs {
		if !existingSet[id] {
			toCreate = append(toCreate, types.Account{
				ID:     id,
				Ledger: Ledger,
				Code:   CodeAgent,
				Flags:  types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
			})
		}
	}

	if len(toCreate) == 0 {
		return nil
	}

	results, err := client.CreateAccounts(toCreate)
	if err != nil {
		return fmt.Errorf("create accounts: %w", err)
	}

	for _, r := range results {
		if r.Result != types.AccountExists {
			return fmt.Errorf("create account index %d: %v", r.Index, r.Result)
		}
	}

	return nil
}
