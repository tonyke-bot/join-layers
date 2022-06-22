# layers.join()

Inspired by [hashlips_art_engine](https://github.com/HashLips/hashlips_art_engine), `layers.join()` is created with Golang to better support generating NFT collectible items efficiently.

PR is open to everyone and let's build the tools for NFT community!

**Core Features**
* YAML-configuration. User-friendly.
* Support PNG files with transparency. More to be added.
* Support multiple layer sets
* High-speed generation by parallelize tasks
* Duplication check
* [WIP] Resume from pause/interrupt

## Example
Thank project [Sweetooth NFT](https://twitter.com/sweetoothnfts/) for providing sample trait files to us to demonstrate how `layers.join()` works to generate a 5K NFT collection.

Please support them on [Sweetooth](https://sweetooth.io/) or [Magic Eden](https://magiceden.io/marketplace/sweetooth).

## Benchmark
Actually I didn't run a serious benchmark for this. But with parallelization, the overall speed is increase a lot comparing to 
other project I've seen on open source community. 

Further speed improvement might be achieved by replacing the image processing with better library. I dunno.

Below test results come from my `Macbook Pro 2021 with M1 Pro (10C16G, 32GB RAM, macOS 12.4, Plugged)`

### Test1: 38s, 600 Collection, 3000x3000, Average Size: 500K. 
```bash
$ join-layers

Metadata folder: <REDACTED>/Desktop/test1/output/json
Image folder: <REDACTED>/Desktop/test1/output/images
Items to generate: 600
Starting workers...
Adding tasks...
Finished! Time used: 38.001s
```

### Test2: 13m52s, 5K Collection, 2048x2048, Average Size: 2.5MB
```bash
$ join-layers

Metadata folder: <REDACTED>/Desktop/test2/output/json
Image folder: <REDACTED>/Desktop/test2/output/images
Items to generate: 5000
Starting workers...
Adding tasks...
Finished! Time used: 13m52.401s
```

## Donation
Any donation to below addresses would be appreciated:
* ETH/BNB: `0x42F30aA6D2237248638D1c74ddfCF80F4ecd340a`
* SOLANA: `2ws1fnQ4U5Gt5361QNEvo6QPuzZSEAUxXoA85Zsdj1wf`
