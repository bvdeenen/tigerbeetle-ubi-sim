/*
Package main is the entry point for the TigerBeetle simulation. It parses CLI flags,
connects to TigerBeetle, bootstraps accounts, and then launches all goroutines —
one per agent, plus the UBI distributor and balance reporter — under a shared context
that is cancelled on SIGINT/SIGTERM.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"

	"github.com/bvdeenen/tigerbeetle-demo/agent"
	"github.com/bvdeenen/tigerbeetle-demo/bank"
	"github.com/bvdeenen/tigerbeetle-demo/reporter"
	"github.com/bvdeenen/tigerbeetle-demo/ubi"
)

func main() {
	nAgents := flag.Int("agents", 5, "number of simulated humans")
	idOffset := flag.Uint64("id-offset", 1000, "first agent account ID")
	tbAddress := flag.String("tb-address", "3000", "TigerBeetle address")
	ubiAmount := flag.Uint64("ubi-amount", 100, "credits issued per UBI cycle per agent")
	ubiInterval := flag.Duration("ubi-interval", 30*time.Second, "how often UBI is distributed")
	minTrade := flag.Uint64("min-trade", 1, "minimum trade amount")
	maxTrade := flag.Uint64("max-trade", 50, "maximum trade amount")
	reportInterval := flag.Duration("report-interval", 10*time.Second, "how often balances are printed")
	flag.Parse()

	if *nAgents < 2 {
		fmt.Fprintln(os.Stderr, "need at least 2 agents")
		os.Exit(1)
	}

	client, err := tb.NewClient(types.ToUint128(0), []string{*tbAddress})
	if err != nil {
		log.Fatalf("connect to TigerBeetle at %s: %v", *tbAddress, err)
	}
	defer client.Close()

	centralBankID := types.ToUint128(1)

	agentIDs := make([]types.Uint128, *nAgents)
	for i := range agentIDs {
		agentIDs[i] = types.ToUint128(*idOffset + uint64(i))
	}

	fmt.Printf("Bootstrapping %d agents (IDs %d–%d)...\n", *nAgents, *idOffset, *idOffset+uint64(*nAgents)-1)
	if err := bank.Bootstrap(client, centralBankID, agentIDs); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	fmt.Println("Accounts ready. Starting simulation.")

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	var wg sync.WaitGroup

	// Reporter
	wg.Add(1)
	go func() {
		defer wg.Done()
		reporter.New(centralBankID, agentIDs, *reportInterval, client).Run(ctx)
	}()

	// UBI distributor
	wg.Add(1)
	go func() {
		defer wg.Done()
		ubi.New(centralBankID, agentIDs, *ubiAmount, *ubiInterval, client).Run(ctx)
	}()

	// Agent goroutines
	for _, id := range agentIDs {
		wg.Add(1)
		id := id
		go func() {
			defer wg.Done()
			agent.New(id, agentIDs, client, *minTrade, *maxTrade).Run(ctx)
		}()
	}

	wg.Wait()
	fmt.Println("Done.")
}
