/*
Package agent implements the simulated human participants. Each Agent runs in its own
goroutine, waking at random intervals to pick a random peer and submit a trade transfer
to TigerBeetle. Agents whose balance would go negative are silently skipped — solvency
is enforced at the database level via the DebitsMustNotExceedCredits account flag.
*/
package agent

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

const (
	ledger    uint32 = 1
	codeTrade uint16 = 1
)

type Agent struct {
	ID       types.Uint128
	peers    []types.Uint128
	client   tb.Client
	minTrade uint64
	maxTrade uint64
	rng      *rand.Rand
}

func New(id types.Uint128, peers []types.Uint128, client tb.Client, minTrade, maxTrade uint64) *Agent {
	seed := id.Bytes()
	s := int64(seed[0]) | int64(seed[1])<<8 | int64(seed[2])<<16 | int64(seed[3])<<24
	return &Agent{
		ID:       id,
		peers:    peers,
		client:   client,
		minTrade: minTrade,
		maxTrade: maxTrade,
		rng:      rand.New(rand.NewSource(s)),
	}
}

func (a *Agent) Run(ctx context.Context) {
	for {
		interval := time.Duration(a.rng.Int63n(int64(a.maxTrade-a.minTrade+1))+int64(a.minTrade)) * time.Second
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			a.attemptTrade()
		}
	}
}

func (a *Agent) attemptTrade() {
	// pick a random peer that isn't self
	var peer types.Uint128
	for {
		peer = a.peers[a.rng.Intn(len(a.peers))]
		if peer != a.ID {
			break
		}
	}

	amount := uint64(a.rng.Int63n(int64(a.maxTrade-a.minTrade+1)) + int64(a.minTrade))

	var debit, credit types.Uint128
	if a.rng.Intn(2) == 0 {
		// I am the buyer: I send money to peer
		debit, credit = a.ID, peer
	} else {
		// I am the seller: peer sends money to me
		debit, credit = peer, a.ID
	}

	results, err := a.client.CreateTransfers([]types.Transfer{
		{
			ID:              types.ID(),
			DebitAccountID:  debit,
			CreditAccountID: credit,
			Amount:          types.ToUint128(amount),
			Ledger:          ledger,
			Code:            codeTrade,
		},
	})
	if err != nil {
		fmt.Printf("[agent %v] transfer error: %v\n", a.ID, err)
		return
	}

	for _, r := range results {
		switch r.Result {
		case types.TransferExceedsCredits,
			types.TransferDebitAccountNotFound,
			types.TransferCreditAccountNotFound:
			// expected: agent is broke or peer not on this instance
		default:
			fmt.Printf("[agent %v] unexpected transfer result: %v\n", a.ID, r.Result)
		}
	}
}
