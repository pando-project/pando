![pando](./documentation/images/pando.jpeg)

[![CircleCI](https://circleci.com/gh/kenlabs/pando/tree/main.svg?style=svg)](https://circleci.com/gh/kenlabs/pando/tree/main)
[![codecov](https://codecov.io/gh/kenlabs/pando/branch/main/graph/badge.svg?token=MFD2QP8RWL)](https://codecov.io/gh/kenlabs/pando)
# Pando
Ensuring access to notarized metadata

## Overview
There are several mechanisms we directly foresee being used for incentivization feedback loops based around the metadata measurements of entities in the filecoin ecosystem. For instance, reports of miner behavior can be used by reputation systems ranking miners, which then are used by clients to select higher quality miners near them. This data doesn’t make sense to directly embed within the filecoin chain for a few reasons: It is produced by independent entities, so the data itself does not need to meet the same ‘consensus’ bar as what we would expect in a global chain, and likewise aspects of reputation and measurements may have aspects of subjectivity. It is also expected that there is diversity of data and that experimentation is a good thing.
However, there are nice properties of having this sort of metadata ecosystem more tightly linked to the chain that seem desirable to encourage, and this leads to the goals for the sidechain metadata service:

* Keep included metadata consistently available
* Provide light-weight, unbiased access to metadata
* Discourage historical revisionism.

## Architecture
![Pando high level architecture](/documentation/images/pando-high-level.png)

## Integration

### As a Provider

As a provider, in order to persist its metadata in Pando service, registration has to be done in advance.

Notice that, any provider needs to integrate a [go-legs publisher](https://github.com/filecoin-project/go-legs/blob/main/publish.go#L20) endpoint to publish its metadata to Pando service.

Please follow the steps below to complete the registration.

1. Download the latest release version of Pando.
2. Decompress the release package.
3. Go to the release directory, execute command `./Pando init` to complete the initialization.
4. Find the `PeerID`, `PrivKey` and `P2PAddr` in `~/.pando/config` for next step.
5. Execute command `./Pando register --privkey <PrivKey> --peerid <PeerID> --pando http://52.14.211.248/ --provider-addr <P2PAddr> --miner <MinerAccount>` to complete the registration. Notice that, if you have no miner account, `--miner` option should be skipped.

### As a Consumer

TBD