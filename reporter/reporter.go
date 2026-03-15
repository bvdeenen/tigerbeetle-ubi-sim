/*
Package reporter periodically fetches all account balances from TigerBeetle and prints
a formatted table to stdout. It is the simulation's only observability mechanism,
showing each account's net balance alongside raw credits and debits posted.
*/
package reporter

import (
	"context"
	"fmt"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Reporter struct {
	CentralBankID types.Uint128
	AgentIDs      []types.Uint128
	Interval      time.Duration
	client        tb.Client
}

func New(centralBankID types.Uint128, agentIDs []types.Uint128, interval time.Duration, client tb.Client) *Reporter {
	return &Reporter{
		CentralBankID: centralBankID,
		AgentIDs:      agentIDs,
		Interval:      interval,
		client:        client,
	}
}

func (r *Reporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.report()
		}
	}
}

func (r *Reporter) report() {
	allIDs := make([]types.Uint128, 0, 1+len(r.AgentIDs))
	allIDs = append(allIDs, r.CentralBankID)
	allIDs = append(allIDs, r.AgentIDs...)

	accounts, err := r.client.LookupAccounts(allIDs)
	if err != nil {
		fmt.Printf("[reporter] lookup error: %v\n", err)
		return
	}

	fmt.Printf("\n[%s] Account Balances\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("  %-20s %10s  %10s  %10s\n", "Account", "Net", "Credits", "Debits")
	fmt.Printf("  %-20s %10s  %10s  %10s\n", "-------", "---", "-------", "------")

	for _, a := range accounts {
		cb := a.CreditsPosted.BigInt()
		db := a.DebitsPosted.BigInt()
		credits := cb.Uint64()
		debits := db.Uint64()

		var net int64
		if credits >= debits {
			net = int64(credits - debits)
		} else {
			net = -int64(debits - credits)
		}

		label := fmt.Sprintf("agent-%v", a.ID)
		if a.ID == r.CentralBankID {
			label = "central-bank"
		}

		fmt.Printf("  %-20s %10d  %10d  %10d\n", label, net, credits, debits)
	}
}
