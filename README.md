# tigerbeetle-ubi-sim

A simulation of a small economy backed by [TigerBeetle](https://tigerbeetle.com) — a
high-performance financial ledger database. Simulated agents buy and sell with each
other at random intervals and amounts, while a central bank periodically distributes a
Universal Basic Income (UBI) to all participants.

## What it does

- **N agents** trade with each other continuously, each waking at a random interval and
  picking a random peer and amount. TigerBeetle enforces solvency — a broke agent's
  transfer is simply rejected at the database level.
- **A central bank** issues UBI to all agents on a fixed interval as an atomic batch.
  The central bank is allowed to run a negative balance (it is the currency issuer).
- **A reporter** periodically prints a balance table to stdout showing each account's
  net balance, total credits, and total debits.

Accounts are created on first run and reused across restarts. Multiple instances can
run in parallel against the same TigerBeetle cluster as long as their `--id-offset`
ranges do not overlap.

## Prerequisites

- Go 1.22+
- A running TigerBeetle instance

```bash
# Create and start a single-node TigerBeetle cluster
tigerbeetle format --cluster=0 --replica=0 --replica-count=1 0_0.tigerbeetle
tigerbeetle start --addresses=3000 0_0.tigerbeetle
```

## Build

```bash
make
```

This produces a `tigerbeetle-ubi-sim` binary in the current directory.

## Usage

```bash
./tigerbeetle-ubi-sim [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--agents` | `5` | Number of simulated agents |
| `--id-offset` | `1000` | TigerBeetle account ID of the first agent |
| `--tb-address` | `3000` | TigerBeetle address |
| `--ubi-amount` | `100` | Credits issued per UBI cycle per agent |
| `--ubi-interval` | `30s` | How often UBI is distributed |
| `--min-trade` | `1` | Minimum trade amount (also minimum sleep between trades, in seconds) |
| `--max-trade` | `50` | Maximum trade amount (also maximum sleep between trades, in seconds) |
| `--report-interval` | `10s` | How often the balance table is printed |

### Example

```bash
./tigerbeetle-ubi-sim --agents=10 --id-offset=1000 --ubi-interval=15s
```

Sample output:

```
Bootstrapping 10 agents (IDs 1000–1009)...
Accounts ready. Starting simulation.
[ubi] distributed 100 credits to 10 agents

[2026-03-15 12:01:00] Account Balances
  Account                     Net     Credits      Debits
  -------                     ---     -------      ------
  central-bank               -1000          0        1000
  agent-3e8                    143        343         200
  agent-3e9                     57        257         200
  agent-3ea                    212        312         100
  ...
```

Press `Ctrl+C` to shut down gracefully.

### Running multiple instances

Each instance manages a non-overlapping range of agent IDs. Use `--id-offset` to
separate them:

```bash
./tigerbeetle-ubi-sim --agents=5 --id-offset=1000 &
./tigerbeetle-ubi-sim --agents=5 --id-offset=2000 &
```

Agents on different instances do not trade with each other, but they share the same
central bank and ledger in TigerBeetle.

## Data model

| | Value |
|--|--|
| Ledger | `1` |
| Central bank account ID | `1` |
| Agent account IDs | `id-offset` … `id-offset + agents - 1` |
| Transfer code — trade | `1` |
| Transfer code — UBI | `2` |
