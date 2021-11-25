Pando API Desgin

## Default Address

```
"Addresses": {
    "Admin": "/ip4/127.0.0.1/tcp/9001",
    "MetaData": "/ip4/127.0.0.1/tcp/9004",
    "GraphSync": "/ip4/127.0.0.1/tcp/9002",
    "DisableP2P": false,
    "P2PAddr": "/ip4/0.0.0.0/tcp/9000",
    "GraphQL": "/ip4/127.0.0.1/tcp/9003"
  },

```

## GraphSync(Go-legs)
	1.localhost:9002/graph/sub/{peerID}
   subscribe a provider(must support go-legs)

## GraphQL
1.localhost:9003
```
eg:
query{
  State(PeerID:"12D3KooWSS3sEujyAXB9SWUvVtQZmxH6vTi9NitqaaRQoUjeEk3M"){
    Cidlist
    LastUpdate
    LastUpdateHeight
  }
  SnapShot(cid:"bafy2bzaced6iroa6ve5m3lxvafcz4ralbuitzpv26f4ks5hznxhqkca5qbxca"){
    Height
    CreateTime
    ExtraInfo
    Update
    PreviousSnapShot
  }
}
```


## SnapShot
    1.localhost:9004/meta/list
List all snapshots’ cids

    2.localhost:9004/meta/info/{cid}
List information about snapshot with cid

    3.localhost:9004/meta/height/{uint}
List information about snapshot with height

## Admin
    1.register [--pando httpaddr] --provider-addr multiaddr --config configpath [or --privkey pkstr -- peerid peeridstr] [--miner mineraccount]
register a provider
```
eg:
register --pando http://52.14.211.248/ --provider-addr /ip4/127.0.0.1/tcp/3102 --config /Users/zxh/config2 --miner t01000

```
pando default addr is localhost
