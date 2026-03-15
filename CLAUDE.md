# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # compile
go run . --agents=5     # run with defaults (TigerBeetle must be running on localhost:3000)
go vet ./...            # lint
```

Starting TigerBeetle locally (single node):
```bash
tigerbeetle format --cluster=0 --replica=0 --replica-count=1 0_0.tigerbeetle
tigerbeetle start --addresses=3000 0_0.tigerbeetle
```

## Architecture

A simulation where N goroutines (agents) buy/sell with each other at random intervals, and a central bank periodically distributes UBI — all backed by TigerBeetle as the ledger.

**Packages:**
- `bank/` — Bootstrap: looks up then creates missing accounts (idempotent, multi-instance safe)
- `agent/` — Per-agent goroutine: picks random peer + amount, creates a transfer, silently ignores broke-agent results
- `ubi/` — Single goroutine: fires on interval, sends linked-batch transfers from central bank to all agents
- `reporter/` — Single goroutine: fires on interval, prints a balance table via `LookupAccounts`
- `main.go` — Wires all components, handles SIGINT via context cancellation

**Data model:**
- Ledger `1`, single currency
- Central bank ID `1` (no spending limit flag — allowed to go negative)
- Agent IDs: `--id-offset + i` (flag `DebitsMustNotExceedCredits` — DB-enforced solvency)
- Transfer code `1` = trade, `2` = UBI

**Key decisions:**
- The shared `tb.Client` is safe for concurrent use across goroutines
- Agent RNGs are seeded per-ID to avoid thundering herd
- UBI transfers use `Linked` flag for atomicity (all-or-nothing per round)
- `types.ID()` (ULID) for transfer IDs — no coordination needed

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--agents` | 5 | Number of simulated agents |
| `--id-offset` | 1000 | First agent account ID |
| `--tb-address` | localhost:3000 | TigerBeetle address |
| `--ubi-amount` | 100 | Credits per UBI cycle per agent |
| `--ubi-interval` | 30s | UBI distribution interval |
| `--min-trade` | 1 | Min trade amount (also used as min sleep seconds) |
| `--max-trade` | 50 | Max trade amount (also used as max sleep seconds) |
| `--report-interval` | 10s | Balance reporting interval |
