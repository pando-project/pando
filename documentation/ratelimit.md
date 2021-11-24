# Rate Limit in Pando

There are 2 types of rate limiter in Pando, gate rate limiter and  peer rate limiter.

The gate rate limiter limits the request rate of all peer, and the peer rate limiters is given different rate limits according to different peer types.

All rate limiter are implemented using token bucket algorithm, and the generation rate of token is:

`k * bandwidth * single DAG size`

This rate is called the `base rate`.

We set `k=0.8` because we considered that it is necessary to reserve at least 20% of the bandwidth for use by other processes on the server.

`bandwidth` is the bandwidth of the environment in which Pando is located. Pando will measure the bandwidth during `pando init` command execution.

`single DAG size` is the size of the data to be transferred of a DAG sync request, default value is `1Mb`

## Table of Contents

- [Gate Limiter](#Gate Limiter)
- [Peer Limiter](#Peer Limiter)
- [Weight of Registered Peer](#Weight of Registered Peer)



## Gate Limiter

The token generation rate of Gate limiter is the base rate which mentioned above. It imposes a rate limit on all graph sync requests. A request must meet two conditions at the same time to be granted: it can get tokens from both Gate limiter and Peer limiter. Otherwise, the request will be paused until it can get tokens from the token buckets above.

![pando rate limit (2)](https://raw.githubusercontent.com/bsjohnson01/resources/master/pando%20rate%20limit%20(2).png)

## Peer Limiter

The rate of Peer limiter  is `m * base rate`, the value of `m` is determined by the type of peer.

We classify peer into three types:

- Unregistered Peer, peer that is not registered with Pando, `m=0.1`
- Whitelist Peer, peer that is configured as a trusted peer in config file, `m=0.5`
- Registered Peer, peer that has been registered with Pando, `m=0.4 * weight`, `weight` is the weight assigend according to the balance of the account provided by the registered provider.



## Weight of Registered Peer

Registered peer will provide its account when registering with Pando. The authenticity of the account will be verified, and Pando will obtain the balance of this account.

The account of peer is classified according to the predefined account balance threshold in the config file. For example, the threshold is defined as follows:

```json
"AccountLevel": {
    "Threshold": [
      1,
      10,
      100,
      500
    ]
  }
```

Then the account level is divided into five levels.

If `account balance ∈ [0, 1)`, the account level is 1, 

`account balance ∈ [1, 10)`, the account level is 2,

...

`account balance ∈[500, +∞]`, the account level is 5

Then we can introduce the formula for calculating the `weight`:

`weight = account level / level count`

`level count` is the number of account level. In this case, its value is 5.

