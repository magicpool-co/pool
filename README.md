# MagicPool Archive

**DISCLAIMER**: IF YOU TRY TO RUN THIS OUT-OF-THE-BOX, IT WILL NOT WORK AND WILL BE VERY FRUSTRATING.
THE INTENTION OF OPEN-SOURCING THIS HAS ALWAYS BEEN AS A REFERENCE TO BUILD YOUR OWN POOL, NOT TO
FORK THE REPO AND RUN THIS CODE AS-IS. DOING SO WILL ONLY CAUSE MINERS TO LOSE EARNINGS
FROM YET ANOTHER POORLY MANAGED POOL. IT IS INEVITABLE PEOPLE WILL TRY, SO LARGE PORTIONS HAVE BEEN
DELIBERATELY EXCLUDED (TERRAFORM CODE, NODE DOCKER IMAGES, ETC). HOPEFULLY, ALONG WITH THE OVERWHELMING
LACK OF DOCUMENTATION, THIS WILL PREVENT AT LEAST SOME OF THE POOR BEHAVIOR THAT COULD COME FROM
OPEN-SOURCING THIS REPO.

This is the pool code for MagicPool, a mining pool that ran in operation for about two and a half
years (shut down September 2023). There is minimal documentation and this is not a support forum
to help you operate a pool. The end-user is meant to be existing or aspiring pool operators
who are looking for a point of reference or an existing implementation for ideas on how
to improve your own pool. 

Operating a pool is not easy and, although I read those exact words before starting this, it is rarely
worth it to run a pool nowadays. Being profitable means having >5-10% of a chains hashrate. Keeping
everything online, handling support requests, and keeping track of protocol changes are only a few
of the daily challenges. It is more than a full-time job that requires extensive technical knowledge.

## Description

The codebase optimizes for high availability (HA) and multi-region support - running it without
this would be both excessive and challenging. This is very different from 
[Open Ethereum Pool](https://github.com/sammy007/open-ethereum-pool), 
[NOMP (original)](https://github.com/zone117x/node-open-mining-portal), or 
[NOMP (Golang)](https://github.com/mining-pool/not-only-mining-pool) for a number of reasons:

  - this is a production pool, without a frontend, that is heavily biased on the infrastructure
    it needs. It is not an open-source, community-driven pool designed to run on any server.
  - MagicPool was always a multi-payout pool (i.e. mine KAS, payout in BTC/ETH), which adds
    a large amount of complexity to all steps after block unlocking.
  - this relies on very few external dependencies - adding new chains means implementing the
    interface properly, along with handling writing transactions and building blocks from scratch.

## High Level Overview

The core of this repo lives in five directories: `core`, `internal`, `pkg`, `app`, and `svc`:
  
  - `core` contains the core components that do many things and have
  	many dependencies (DB, Redis, nodes, etc.). These are not applications, but applications
  	do use them. Examples include `core/trade` (exchange handler), `core/payout` (payout handler),
  	and `core/credit` (block unlocker/creditor).
  - `internal` contains I/O handlers (with the exception of `internal/accounting`). These are
  	things like `internal/pooldb` (MySQL pool database, with migrations, reads, and writes) and
  	`internal/node` (full nodes). `internal/accounting` handles the math for both round (block crediting)
  	and exchange (BTC/ETH payout) accounting.
  - `pkg` contains general purpose packages that have no place elsewhere. Examples include
  	`pkg/aws` (handling for `Route53`, `SES`, `EC2`, etc.), `pkg/crypto` (implements transaction/block building
  	along with other helpers like base58 and mnemonics), and `pkg/stratum` (implements stratum server + client, with some
  	bias towards the needs of the MagicPool).
  - `app` contains the actual application code for the `pool`, `worker`, and `api`. `pool` is the actual pool
  	(a glorified TCP server), `worker` is a cron handler with distributed locks, and `api` is a standard lib API router 
  	and HTTP server for the frontend.
  	This directory contains the heavy lifting for the applications, referencing all three of the above directories.
  - `svc` contains the services and Dockerfiles to build each application. `svc/runner.go` is a service runner
    (designed for AWS ECS) that handles SIGTERM and SIGINT in a fairly graceful way. `svc` and `app` are separate
    for the sake of integration testing (in `tests/`), because if each application in `app/` was a `main` package,
    integration testing becomes very challenging.

## Design Choices

 - the goal has always been to share the pool code for all different chains. In Go, this means having some interface
   that the pool interacts with. The challenge with this approach is that often times it is more trouble than it's worth
   when two chains act very differently. We did our best to mitigate this by doing things like having each chain provide
   its own `mining.subscribe`, `mining.authorize`, and `mining.submit` responses, but often times it feels very bulky. If
   we were to do this again, it would probably be a mistake to force all chains into a single interface (unless all chains
   are extremely similar).
 - the safe assumption is always that Redis data could disappear at any moment. We designed this in a way where, if Redis did
   get cleared, the only data that would really matter would be the PPLNS window. Every time a block is found,
   all necessary data (block hash, miner shares, etc) are immediately stored in the database.
 - full nodes failing is the other main risk. To mitigate this, we created a "hostpool" (`pkg/hostpool`) that acts as an
   intelligent reverse proxy for nodes (HTTP, TCP, and GRPC supported). So you would run a full node (per chain) in every region,
   and the hostpool automatically uses the one with the lowest latency. If that node goes down, it will automatically switch over
   to the node with the next lowest latency. For requests that need to be sticky (i.e. block submission, since work packages
   exist only on one specific node), hostpool supports forcing a request to that node.
 - after sending duplicate payouts, we made a serious effort to make sure a specific payout only happens once. This happens by
   separating the action of initiating transactions and sending transactions into two services (cron jobs). By pre-specifying
   the UTXOs to use (for UTXO-based chains) and storing the full signed transaction in the database prior to sending, you can
   eliminate duplicate transactions (since non UTXO-based chains still rely on a nonce, which increments on each tx). Ideally
   the full transactions would be stored unsigned, but certain chains (mainly ERG) are very challenging to separate the signing
   step when spending block rewards, so couldn't be accomplished.
 - we attempted to manage our own wallets and UTXO-sets for each chain instead of running node-based wallets. This works 
   for every one except for ERG and we were always grateful that we took this approach. It is far more secure and 
   allows for a lot more flexibility, especially when you have to do things like merge UTXO-sets and decide who pays the 
   transaction fees. It does mean you need native transaction and wallet support for each chain, all of which lives in `pkg/crypto`.
 - BTC/ETH payouts are complex. The two main points are whether or not the exchange is allowing deposits/trades/withdrawals for
   the specified chain and whether or not the cumulative BTC/ETH surpasses your set withdrawal threshold (to minimize 
   withdrawal fees while maximizing the frequency of exchange batches). Implementing each exchange (whether they have
   market orders or quote initial quantity, handling trade vs main accounts, different rounding techniques,
   different fee structures, etc.) is an ordeal - much of the work we did was heavily inspired by
   [CCXT](https://github.com/ccxt/ccxt). One strategy (that we didn't implement) to minimize price fluxuation would be to 
   periodically (say, every hour) sending coins to the exchange and making the trades then. For every chain we mine,
   the transaction fees are miniscule, and this would mean you only have an hour of price fluxuation. The downside is exchange
   risk, which we always were adverse to. We decided to minimize the time on exchange (instead of minimizing price fluxuation)
   with the assumption that, at scale, we would be making enough batches per day to overcome the fluxuation.

## Infrastructure

All of our infrastructure was on AWS for a number of reasons, the main being assurances of security and uptime. You can
cut costs hugely by running it on bare metal but your day-to-day will become that much more complex. All of our infrastructure
was managed through Terraform since, at this number of resources, manual management is untenable 
(at peak it was over 1200 resources, ~50 full nodes across 6 different regions alone). The main resources
we used are as follows:

  - Elastic Container Service (ECS) to run full nodes, the pools, the workers, and the APIs. Kubernetes costs $80/mo
  	per cluster, which adds significant overhead if you're running in multiple regions (and single-tenant would be cost
  	prohibitive).
  - Global Aurora (MySQL) for our database. We used a single master, with read-only replicas in each API region. 
  	This is quite expensive (our largest cost), but API latency is just too great if you only do a single region. 
  	We also implemented a rough version of auto-failover in the case of the master's region going down.
  - Global Elasticache (Redis) for our pool and API cache. This has native auto-failover and is far cheaper than Global Aurora.
  - CI/CD through Github Actions (self-hosted runners for ARM images) that builds Docker images and rolls out a new deployment in ECS.
  - self-hosted Loki, Prometheus, and Grafana for our logging and metrics.

If we were to do this over again, I would use AWS without a doubt. The burden of a pool without having to constantly
service infrastructure is enough, bare metal would require a larger team (that would be more expensive than the cost
differential).

## Chain Notes

Chains vary in how easy they are to implement (or how similar they are to another chain). We developed 
[powkit](https://github.com/sencha-dev/powkit) to keep all hash functions natively in Go, which helped in
some aspects, but the difference in stratum protocols, work packages, and node durability can make things
very challenging. These are some rough notes on the pros and cons of each chain we implemented:

  - *Aeternity*: An account-based chain, similar to ETH, that builds work packages within the node. This
  	isn't a horrible chain to implement, but it does have two major challenges: a custom version of RLP encoding
  	for transactions, and no call to get the block reward for a given block. This means a fairly custom transaction
  	building (they do have a [Go SDK](https://github.com/aeternity/aepp-sdk-go)) and making external HTTP calls to
  	get block reward. The node is fairly resilient, no real issues (ARM builds for Docker were an ongoing issue at
  	the time, unsure if it is solved now).
  - *Conflux*: An account-based chain, similar to ETH, that uses its own internal Stratum from the node. Conflux is
  	not fun to implement. Maintaining the TCP connection to the node is very annoying and feels excessive. The node
  	also prunes data excessively (often times after 24 hours), the node requires a huge amount of storage (500GB minimum
  	at the time of writing), and takes hours to start up after a restart. We had to call their external archival node for
  	certain calls after 24hr, which often times rate limits and may not even be public anymore. We stopped running Conflux
  	after a few days.
  - *Cortex*: An ETH clone with some weird changes (namely cuckoo as the proof-of-work function). This is a dead chain
  	that rarely syncs, and is not even worth going into depth. All RPC calls are the same as ETH/ETC, which is nice, but
  	this is a waste of time if you want any assurances or consistency.
  - *Ergo*: A UTXO-based chain that is fairly unique. Implementing this in Rust would be best, because you could leverage
  	[sigma-rust](https://github.com/ergoplatform/sigma-rust). Without a native library like that, you have to heavily rely
  	on the node's HTTP calls which can be strange (and means managing the wallet inside of the node). Spending block rewards
  	also is a unique process that has to be done a specific way. The node is fairly stable, but the UTXO index can get corrupted
  	after a few months and requires a re-index. Node was very stable, ARM Docker builds are iffy though.
  - *Ethereum Classic (+variants)*: An ETH clone only differing in the proof-of-work function (minor variations to Ethash constants).
  	Pretty reasonable to run, very durable node, no real complaints (other than general ETH complaints).
  - *Firo*: A BTC-like chain that uses Firopow, a Progpow variant. In general pretty similar to other Progpow chains, but the node
  	often had bugs in the RPC responses that could cause some problems. Changes in the dev fees are fairly annoying to keep track of
  	since they are hardcoded. Dev team/support is minimal. Node is very stable once you get it operating properly, minimal 
  	datadir size.
  - *Flux*: A ZEC fork that removes shielded transactions. It uses a modified version of Equihash that they call Zelhash. The only
  	real change is a "twisting step" (more info in `powkit` repo). A fairly reasonable chain to run. I often wondered why it was a
  	ZCash fork instead of a Bitcoin fork since most of the ZCash features were removed. Signing transactions is the "ZCash way",
  	which is very annoying to implement natively so we leveraged the full node's signing RPC call for simplicity.
  - *Kaspa*: A UTXO-based chain that is very unique. It uses a DAG structure that is fundamentally different from a blockchain. The
  	block rate is very high (1 per second) and will increase to be far higher (10 per second?) when the rust-rewrite is finished.
  	Pruning is similar to Conflux, except three days of history are stored. We operated an archive node though after block time
  	increases, that will be untenable. Block unlocking and transaction completion are very challenging to implement properly, but
  	the dev team is helpful with any questions. It uses a modified version of BTC transactions so that has to be written natively.
  	It is not easy to do a Kaspa pool properly and it's safe to assume most pools are losing blocks. IceRiver ASICs are ridden with
  	bugs (main one is that often times it sends solutions for old work with the wrong job ID, ~5% of the time). Most pools have
  	a high rejected share rate because they aren't working around this.
  - *Nexa*: A UTXO-based chain that is a BCH fork. It has lots of changes, namely the proof-of-work function and transaction
  	building/signing. It is very stable once you get it going. It uses BCH-style Schnorr signatures (different from BTC Schnorr)
  	which were implemented natively. Block maturity period is seven days which can cause some confusion.
  - *Ravencoin*: A BTC-like chain that uses Kawpow, a Progpow variant. In general pretty similar to other Progpow chains. No real
  	complaints with this, if you've implemented other BTC chains it's mostly like that.

