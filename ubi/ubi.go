/*
Package ubi implements the central bank's universal basic income distribution. On each
interval it submits a single linked batch of transfers — one per agent — debiting the
central bank and crediting every agent atomically. If any transfer in the batch fails,
the entire round is rolled back and retried next interval.
*/
package ubi

import (
	"context"
	"fmt"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

const (
	ledger  uint32 = 1
	codeUBI uint16 = 2
)

type Distributor struct {
	CentralBankID types.Uint128
	AgentIDs      []types.Uint128
	Amount        uint64
	Interval      time.Duration
	client        tb.Client
}

func New(centralBankID types.Uint128, agentIDs []types.Uint128, amount uint64, interval time.Duration, client tb.Client) *Distributor {
	return &Distributor{
		CentralBankID: centralBankID,
		AgentIDs:      agentIDs,
		Amount:        amount,
		Interval:      interval,
		client:        client,
	}
}

func (d *Distributor) Run(ctx context.Context) {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.distribute()
		}
	}
}

func (d *Distributor) distribute() {
	transfers := make([]types.Transfer, len(d.AgentIDs))
	for i, agentID := range d.AgentIDs {
		flags := types.TransferFlags{}
		if i < len(d.AgentIDs)-1 {
			flags.Linked = true
		}
		transfers[i] = types.Transfer{
			ID:              types.ID(),
			DebitAccountID:  d.CentralBankID,
			CreditAccountID: agentID,
			Amount:          types.ToUint128(d.Amount),
			Ledger:          ledger,
			Code:            codeUBI,
			Flags:           flags.ToUint16(),
		}
	}

	results, err := d.client.CreateTransfers(transfers)
	if err != nil {
		fmt.Printf("[ubi] error: %v\n", err)
		return
	}

	if len(results) > 0 {
		fmt.Printf("[ubi] round failed (first error index=%d result=%v)\n", results[0].Index, results[0].Result)
		return
	}

	fmt.Printf("[ubi] distributed %d credits to %d agents\n", d.Amount, len(d.AgentIDs))
}
