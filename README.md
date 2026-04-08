# DisPAGraph

> **Dis**tribution **P**latform for **A**nonymized Knowledge **Graph**s

A traceable, privacy-preserving, blockchain-based platform for distributing anonymized knowledge graphs. Built on **Hyperledger Fabric 2.2**, written in **Go**, benchmarked with **Hyperledger Caliper**.

_Master's Thesis — Università degli Studi dell'Insubria, Computer Science, A.Y. 2021-2022_

---

## Table of Contents

- [Overview](#overview)
- [Financial Industry Use Case](#financial-industry-use-case)
- [System Architecture](#system-architecture)
- [Privacy Model](#privacy-model)
- [Smart Contracts](#smart-contracts)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Running Caliper Benchmarks](#running-caliper-benchmarks)
- [Performance Results](#performance-results)
- [Project Structure](#project-structure)
- [Security Properties](#security-properties)

---

## Overview

Modern organizations represent interconnected data as **Knowledge Graphs (KGs)**, directed, labelled graphs that capture both attribute values and relationships between entities. While KGs are powerful, sharing them exposes users to re-identification attacks: an adversary can correlate attributes and graph topology to de-anonymize individuals even when explicit identifiers are removed.

**DisPAGraph** solves this problem by providing a blockchain-based platform where:

- **Data providers** collect personal data from users, anonymize it into a KG, and distribute it to requesting parties — with a cryptographic proof that every user's privacy preference is respected.
- **Data owners** share personal information encrypted end-to-end, specify their own privacy strength (_k_ value), and receive a zero-knowledge **Proof-of-Privacy** guaranteeing their _k_ was honoured in any shared KG.
- **Data recipients** request anonymized KGs knowing the blockchain ledger provides an immutable, auditable record of every data exchange.
- A **Rollup system** (trusted off-chain verifier) evaluates the anonymized KG and generates the Proof-of-Privacy using **Pedersen commitments** over the Edwards25519 elliptic curve.

Anonymization is performed with the **Personalized _k_-Attribute Degree (k-ad)** model (Hoang et al., ACNS 2020), an extension of k-anonymity for KGs that lets every user set their own re-identification confidence threshold of 1/k.

---

## Financial Industry Use Case

> **Scenario: Inter-bank Credit Risk Knowledge Graph Sharing**

Financial institutions maintain rich KGs of customer transaction behaviour — nodes represent customers and account attributes (income bracket, credit score, transaction volume); edges represent relationships (transfers, co-signatories, guarantors). These graphs are invaluable for credit scoring, fraud detection, and AML compliance — but sharing them with credit bureaus or regulators raises severe GDPR and data-privacy concerns.

**How DisPAGraph addresses this:**

```
┌─────────────────────────────────────────────────────────────────┐
│  Bank A (Data Provider)                                         │
│  · Creates a campaign: "Q3 Credit Risk Assessment"              │
│  · Collects encrypted customer KG data on-chain                 │
│  · Generates anonymized KG respecting each customer's k value   │
│  · Rollup system verifies: every customer is indistinguishable  │
│    from ≥ k-1 others in the anonymized graph                    │
│  · Stores Proof-of-Privacy on the Fabric ledger                 │
└───────────────────────┬─────────────────────────────────────────┘
                        │  Fabric channel (immutable audit log)
          ┌─────────────┴──────────────┐
          │                            │
┌─────────▼──────────┐      ┌──────────▼────────────┐
│ Credit Bureau      │      │ Regulator (e.g. ECB)   │
│ (Data Recipient)   │      │ (Data Recipient)        │
│ · Requests KG      │      │ · Requests KG           │
│ · Verifies proof   │      │ · Independently audits  │
│ · Decrypts KG      │      │   the proof on-chain    │
└────────────────────┘      └───────────────────────-┘
```

**Key regulatory benefits:**

| Requirement | How DisPAGraph Satisfies It |
|---|---|
| GDPR Art. 25 — Privacy by Design | Personalized k-ad anonymization; each customer sets their own k |
| GDPR Art. 5(f) — Integrity & Confidentiality | ElGamal ECC digital envelopes; only intended recipient can decrypt |
| MiFID II / DORA — Auditability | Every data exchange recorded immutably on Hyperledger Fabric ledger |
| AML — Data lineage | Blockchain traces exactly which institution received which KG version |
| SOX / Basel III — Non-repudiation | ElGamal digital signatures; rollup system signs every proof |

**Concrete example:** Bank A needs to share transaction graph data with a credit bureau for loan underwriting. Under legacy approaches the bureau receives a raw CSV; under DisPAGraph:

1. Customer Alice specifies `k=5` — she must be indistinguishable from 4 others.
2. Customer Bob specifies `k=10` — higher privacy for sensitive profile.
3. Bank A runs Personalized k-ad and obtains anonymized KG G̅.
4. The rollup system verifies G̅, generates `POP(Comm, Comm⁺)` and signs it.
5. The smart contract checks `Comm == Comm⁺` — proves Alice ≥ k=5, Bob ≥ k=10.
6. The credit bureau receives G̅ encrypted with their public key; only they can decrypt.
7. The entire exchange is permanently auditable on the blockchain.

Alice and Bob can independently query the ledger at any time to verify the proof — without ever seeing the actual KG.

---

## System Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                    Hyperledger Fabric Network                       │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │  obs0 (Peer) │  │  prov0 (Peer)│  │  rec0 (Peer) │             │
│  │  Data Owners │  │  Providers   │  │  Recipients  │             │
│  │              │  │              │  │              │             │
│  │  S1 S2 S3    │  │  S1 S2 S3    │  │  S1 S2 S3    │             │
│  │  Ledger L1   │  │  Ledger L1   │  │  Ledger L1   │             │
│  │  CouchDB     │  │  CouchDB     │  │  CouchDB     │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘             │
│         └─────────────────┴──────────────────┘                     │
│                           │  Channel C1                            │
│                    ┌──────┴──────┐                                 │
│                    │  Orderer    │                                  │
│                    └─────────────┘                                 │
└────────────────────────────────────────────────────────────────────┘
           │                                        │
┌──────────▼──────────┐                  ┌──────────▼──────────┐
│  Data Provider      │                  │  Rollup System       │
│  · Runs k-ad anon.  │◄────────────────►│  · Stores KGs        │
│  · Manages campaign │                  │  · Verifies KGs      │
└─────────────────────┘                  │  · Generates POP     │
                                         └──────────────────────┘
```

**Organizations:**

| Organization | Domain | Role |
|---|---|---|
| `obs0` | `obs0.tracenet.com` | Data owners — share personal data |
| `prov0` | `prov0.tracenet.com` | Data providers — anonymize & distribute KGs |
| `rec0` | `rec0.tracenet.com` | Data recipients — consume anonymized KGs |

Each peer hosts three chaincodes (`S1=campaign`, `S2=ownerData`, `S3=anonymizedKG`) and a CouchDB ledger instance for rich JSON queries.

---

## Privacy Model

DisPAGraph implements the **DisPAGraph Proof-of-Privacy** — a zero-knowledge proof scheme built on **Pedersen commitments** over Edwards25519.

Given data owners O = {o₁, ..., oₙ} with privacy preferences K = {k_o1, ..., k_on} and their actual privacy values K' = {k'_o1, ..., k'_on} in anonymized KG G̅, the proof demonstrates in zero knowledge that:

```
∀ oᵢ ∈ O :  k'_oi ≥ k_oi
```

**Proof generation (rollup system + data owners):**
1. Rollup generates random blinding factor r_oi for each owner.
2. Each owner oᵢ computes `Comm_oi = Comm(k'_oi - k_oi, r_oi)` and sends it back.
3. Rollup aggregates: `Comm = Σ Comm_oi`.

**Proof verification (smart contract):**
1. Rollup computes `Comm⁺_oi = Comm(|k'_oi - k_oi|, r_oi)` for each owner.
2. Aggregates: `Comm⁺ = Σ Comm⁺_oi`.
3. Smart contract checks `Comm == Comm⁺` — true iff k'_oi ≥ k_oi for all owners.

The proof reveals nothing about the individual k values — only that the constraint holds.

---

## Smart Contracts

Three chaincodes implement the full DisPAGraph workflow:

### `campaign` — Campaign Management

```go
CreateCampaign(id, name, startTime, endTime string) error
QueryCampaign(id string) (*Campaign, error)
CampaignExists(id string) (bool, error)
DeleteCampaign(id string) error
```

A **campaign** is a data collection event created by a provider. Data owners discover open campaigns and choose which to contribute to.

### `ownerData` — Data Owner Sharing

```go
ShareData(id, campaignId, envelope, privacyPreference string) error
DeleteSharedData(id string) error
```

Stores the encrypted data envelope (ElGamal digital envelope of the symmetric key) on-chain. The rollup system uses this to retrieve and decrypt the data owner's personal information. Validates that the campaign exists and has not ended.

Cross-chaincode invocation: calls `CampaignExists` on the `campaign` chaincode.

### `anonymizedKG` — KG Lifecycle

```go
StoreAnonymizedKG(id, campaignId, recipientId, rollupEnvelope, signature string) error
StoreProof(KGId, userCommit, rollupCommit string) (bool, error)
ShareAnonymizedKGWithRecipient(KGId, campaignId, recipientId, recipientEnvelope string) error
DeleteAnonymizedKG(id string) error
```

Manages the full lifecycle of an anonymized KG:
1. **Store** — provider submits the KG identifier and encrypted envelope for the rollup system.
2. **StoreProof** — rollup system submits `(Comm, Comm⁺)`; contract verifies equality and marks KG as `Verified`.
3. **Share** — provider shares the KG with the recipient, only allowed after verification.

The `Caliper*` variants (`CaliperStoreProof`, `CaliperShareAnonymizedKGWithRecipient`) write to unique dummy IDs to avoid MVCC conflicts in concurrent benchmark workloads.

---

## Prerequisites

| Dependency | Version | Purpose |
|---|---|---|
| [Go](https://go.dev) | ≥ 1.19 | Chaincode language |
| [Docker](https://www.docker.com) | ≥ 20.x | Container runtime |
| [Docker Compose](https://docs.docker.com/compose/) | ≥ 2.2 | Network orchestration |
| [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-2.2/install.html) | 2.2 | Blockchain platform |
| [Node.js / npm](https://www.npmjs.com) | ≥ 14.x | Caliper benchmarks |
| [jq](https://stedolan.github.io/jq/) | ≥ 1.6 | JSON processing in scripts |

---

## Quick Start

**1. Set script permissions**

```bash
sudo chmod 755 main.sh settings.sh
sudo chmod -R 755 scripts/
```

**2. Start the network**

```bash
./main.sh network restart
```

This single command:
- Tears down any existing network and cleans state
- Generates crypto material with `cryptogen`
- Starts Docker Compose (orderer + 3 peers + 3 CouchDB instances)
- Creates channel `mychannel`
- Joins all three organizations to the channel
- Sets anchor peers
- Packages, installs, approves, and commits all three chaincodes

**3. Verify the network is running**

```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

You should see `orderer`, `peer0.obs0`, `peer0.prov0`, `peer0.rec0`, and three CouchDB containers all healthy.

**4. Invoke a smart contract manually**

```bash
# Create a campaign
./scripts/chaincodeOperation.sh createCampaign campaign1 "Q3 Risk Assessment" 1700000000 1800000000

# Query it back
./scripts/chaincodeQuery.sh queryCampaign campaign1
```

---

## Running Caliper Benchmarks

DisPAGraph ships with a complete Hyperledger Caliper benchmark suite covering all five smart contract operations and all three blockchain query types.

**Initialize Caliper**

```bash
./main.sh caliper init
```

**Run benchmarks**

```bash
./main.sh caliper launch campaign          # CreateCampaign
./main.sh caliper launch shareData         # ShareData
./main.sh caliper launch storeAnonymizedKG # StoreAnonymizedKG
./main.sh caliper launch storeProof        # StoreProof
./main.sh caliper launch shareAnonyKG      # ShareAnonymizedKG
```

Each benchmark sends **2,000 transactions** per round across **10 send-rate levels** (10–100 TPS), after initializing the ledger with 2,000 transactions. Results are written to `caliper/report.html`.

---

## Performance Results

Benchmarks were executed on the following hardware:

| Component | Specification |
|---|---|
| CPU | Intel Core i5-7200U @ 2.50 GHz |
| RAM | 8 GB |
| OS | Ubuntu 20.04 LTS |
| Language | Go 1.19.2 linux/amd64 |
| Containerization | Docker Compose v2.2.3 |
| Commitment library | `bwesterb/go-ristretto` (Edwards25519) |

### Smart Contract Throughput

Smart contracts were grouped by ledger access pattern:

| Group | Operations | Access Pattern |
|---|---|---|
| **Group A** | `CreateCampaign`, `StoreProof`, `ShareAnonymizedKG` | 1 read + 1 write |
| **Group B** | `ShareData`, `StoreAnonymizedKG` | 2 reads + 1 write (cross-chaincode) |

| Send Rate (TPS) | Group A Throughput (TPS) | Group A Latency (s) | Group B Throughput (TPS) | Group B Latency (s) |
|:---:|:---:|:---:|:---:|:---:|
| 10 | 10 | < 0.5 | 10 | < 0.5 |
| 20 | 20 | < 0.5 | 19 | < 0.5 |
| 30 | 30 | < 0.5 | 29 | < 0.5 |
| 40 | 39 | 1.1 | 38 | 2.2 |
| **50** | **~47 (peak)** | 2.1 | 40 | 3.5 |
| 60 | 43 | 3.2 | 40 | 4.1 |
| 70 | 42 | 3.8 | 38 | 4.5 |
| 80 | 41 | 4.1 | 37 | 4.8 |
| 90 | 40 | 4.5 | 36 | 5.0 |
| 100 | 39 | 4.8 | 36 | 5.1 |

> Group A peaks at **~47 TPS** (send rate 50); Group B saturates at **~40 TPS** due to cross-chaincode reads adding verification overhead.

### Query Throughput

Read-only queries (no consensus required) significantly outperform write transactions:

| Query | Purpose |
|---|---|
| `RetrieveEnvelopeRollupSystem` | Rollup retrieves data owner's encrypted envelope |
| `RetrieveKGEnvelopeRollupServer` | Rollup retrieves provider's KG envelope |
| `RetrieveKGEnvelopeRecipient` | Recipient retrieves provider's encrypted KG |

| Send Rate (TPS) | Throughput (TPS) | Latency (s) |
|:---:|:---:|:---:|
| 10 | 10 | < 0.1 |
| 20 | 20 | < 0.1 |
| 30 | 30 | < 0.1 |
| 40 | 40 | < 0.1 |
| 50 | 50 | < 0.1 |
| 60 | 58 | < 0.1 |
| **70** | **~70 (peak)** | **0.53** |
| 80 | 60 | 4.0 |
| 90 | 60 | 5.5 |
| 100 | 60 | 6.8 |

> Queries sustain **~60–70 TPS** with sub-second latency — well above the write-transaction ceiling.

### System-Level Capacity

In a real deployment, one complete "user data share" requires **1 transaction + 1 query**, and one complete "recipient KG request" requires **3 transactions + 2 queries**. Given the measured peak of ~50 operations/second:

- The system can serve approximately **16 concurrent user operations per second**.
- Extrapolated to continuous 24-hour operation: **>1 million user interactions per day**.

---

## Project Structure

```
tracenet/
├── chaincode/
│   ├── campaign/            # Campaign management chaincode (Go)
│   │   ├── campaign.go      # CreateCampaign, QueryCampaign, CampaignExists
│   │   └── main.go
│   ├── ownerData/           # Data owner sharing chaincode (Go)
│   │   ├── ownerData.go     # ShareData, DeleteSharedData
│   │   └── main.go
│   └── anonymizedKG/        # KG lifecycle chaincode (Go)
│       ├── anonymizedKG.go  # StoreAnonymizedKG, StoreProof, ShareAnonymizedKGWithRecipient
│       └── main.go
├── caliper/
│   ├── benchmarks/          # Caliper YAML benchmark configs (10–100 TPS)
│   │   ├── campaign.yaml
│   │   ├── shareData.yaml
│   │   ├── storeAnonymizedKG.yaml
│   │   ├── storeProof.yaml
│   │   └── shareAnonymizedKG.yaml
│   ├── workload/            # Caliper JavaScript workload modules
│   │   ├── createCampaign.js
│   │   ├── shareData.js
│   │   ├── storeAnonymizedKG.js
│   │   ├── storeProof.js
│   │   ├── shareAnonymizedKG.js
│   │   └── retrieveEnvelope.js
│   └── networks/
│       └── tracenet-config.yaml  # Caliper network topology config
├── config/
│   ├── configtx.yaml        # Channel & policy configuration
│   ├── crypto-config.yaml   # Cryptogen certificate generation
│   └── core.yaml            # Peer configuration
├── docker/
│   ├── docker-compose.yml   # Full network topology (3 peers, 3 CouchDB, orderer)
│   └── docker-compose-base.yml
├── scripts/
│   ├── init.sh              # Crypto material + genesis block generation
│   ├── network.sh           # Docker Compose lifecycle
│   ├── channel.sh           # Channel creation, join, anchor peer setup
│   ├── deployChaincode.sh   # Package, install, approve, commit chaincodes
│   ├── chaincodeOperation.sh
│   ├── chaincodeQuery.sh
│   ├── caliper.sh
│   └── utils/
│       ├── environment.sh
│       ├── output.sh
│       └── connectionProfile.sh
├── main.sh                  # Master orchestration entry point
└── settings.sh              # Environment variables (ports, versions, channel name)
```

---

## Security Properties

| Property | Mechanism |
|---|---|
| **Confidentiality** | ElGamal ECC digital envelopes — only the intended recipient can decrypt the symmetric key |
| **Integrity** | Hyperledger Fabric immutable ledger — transactions cannot be altered once committed |
| **Authenticity** | ElGamal digital signatures — rollup system signs every Proof-of-Privacy |
| **Accountability** | Pedersen commitment Proof-of-Privacy — data owners can verify their k was respected without learning anything else about the KG |
| **Non-repudiation** | Blockchain audit trail — every data exchange (who shared what with whom, when) is permanently recorded |
| **Anti-replay** | Smart contracts enforce unique IDs and single-share semantics — a recipient cannot receive more than one snapshot of the same KG |

---

## References

- A.-T. Hoang, B. Carminati, E. Ferrari. _"Cluster-based anonymization of knowledge graphs."_ ACNS 2020.
- A.-T. Hoang, B. Carminati, E. Ferrari. _"Personalized anonymization of knowledge graphs."_ Under review, 2023.
- E. Androulaki et al. _"Hyperledger Fabric: a distributed operating system for permissioned blockchains."_ EuroSys 2018.
- T. P. Pedersen. _"Non-interactive and information-theoretic secure verifiable secret sharing."_ CRYPTO '91.
- T. ElGamal. _"A public key cryptosystem and a signature scheme based on discrete logarithms."_ CRYPTO 1985.
