# GroupFi Proxy API

## url 
```
https://api.groupfi.ai
```

## public apis
### GET /mint_nicknft, mint a nft which contains user's nick name
#### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|address|string| user's address, smr address|
|name|string|letters and numbers, lowercase, 8 <= length <= 20|
|image| string, optional|image url|
#### response
```json
{
    "result": true
}
```

### GET /smr_price
#### response
```json
{
    "result" : true,
    "data" :{
        "chain1":{
            "contract" : "0x11",
            "token": "eth",
            "price" : "100000",
            "deci": 18
        },
        "chain1":{
            "contract" : "0x22",
            "token": "bsc",
            "price" : "100000",
            "deci": 18
        }
    }
}
```

### POST /group/filter, filter the addresses, remove ones don't belong to the group
#### params
```json
{
    "chain": 148,
    "addresses" : ["address1","address2"],
    "contract" : "group contract address",
    "threshold" : "1000000000",
    "erc" : 721,
    "ts" : 1712647238
}
```
erc is 20 or 721.
#### response
```json
{
    "result": true,
    "indexes": [],
}
```

### POST /group/verify, verify the addresses that belong to one group or not
#### params
```json
{
    "chain": 148,
    "adds" : ["address1","address2"],
    "subs" : ["address1","address2"],
    "contract" : "group contract",
    "threshold" : "100000",
    "erc" : 20,
    "ts" : 1712647238
}
```
#### response
```json
{
    "result": true,
    "flag":   1,
    "sign": "sign hex string"
}
```

### GET /proxy/account?publickey=0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27, get the proxy account using user's pairX 
#### response
```json
{
    "result": true,
    "proxy_account": "a smr address"
}
```

## sign apis
### POST /proxy/register, register a new proxy account or update the sign_account
#### params
```json
{
    "encryptedPrivateKey": "0x7b2276657273696f6e223a227832353531392d7873616c736132302d706f6c7931333035222c226e6f6e6365223a226148496b55726e724a68534c374b445a655057415a707658554c6442495a4344222c22657068656d5075626c69634b6579223a226d4c566e6f2b4943326776417641342f766330516639656638362f65394c5a6377562f34576b56565a6d513d222c2263697068657274657874223a224f477236757734756c6c66465a74712b762b65707a302b4f4c4738714879324e754347445959593078726f6b7464304d763251326646497678527136505041397573353666304e2f4169446a4e6172775768416764486e4d546f312b65796b39536b712b633841635665633d227d",
    "pairXPublicKey": "0x9b3da30c3aa890958b95e96b65a5e0f77a28cb1211d097ab943ef03d9dab9651",
    "evmAddress": "0x928100571464c900A2F53689353770455D78a200",
    "timestamp": 1711449778,
    "scenery": 1,
    "signature": "0xccec1e146ff48198566e706d548536c4cc3e6afa3ac351c740fb9f951912b90f1fb064f33682ac12f9e9fad446e3a9dc7ce53dd81c36729fa41cf946f4d1138c1b"
}
```
#### response
```json
{
    "result": true,
    "proxy_account": "a smr address"
}
```

### POST /proxy/mint_nicknft, mint a nft which contains user's nick name for the proxy address
#### params
```json
{
    "publickey": "0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
    "data": "name string, letters and numbers, lowercase, 10 <= length <= 20",
    "ts": 1711449778,
    "sign": "0x0000000000000000000000, sign(ts), using the sign_account"
}
```
#### response
```json
{
    "result": true
}
```



### POST /proxy/send, send the TransactionEssence which contains a msg as metadata on shimmer L1 network
#### params
```json
{
    "publickey": "0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
    "data":"0x0, TransactionEssence, bytes",
    "ts": 1711449778,
    "sign": "0x0000000000000000000000, sign(data+ts), using the sign_account"
}
```
#### response
```json
{
    "result": true,
    "transactionid": "transaction id",
    "blockid": "bock id"
}
```

### POST /proxy/send/asyn, send the TransactionEssence which contains a msg as metadata on shimmer L1 network, asynchronization
#### params
```json
{
    "publickey": "0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
    "data":"0x0, TransactionEssence, bytes",
    "ts": 1711449778,
    "sign": "0x0000000000000000000000, sign(data+ts), using the sign_account"
}
```
#### response
```json
{
    "result": true,
    "transactionid": "transaction id",
    "blockid": "bock id"
}
```

## Error Response
### Response format
```json
{
    "result": false,
    "err-code": 0,
    "err-msg": "system error"
}
```

### ERROR CODE
|code | description|
|---|------------|
|1|the signed ts is time out, 10 minutes|
|2|sign error, can not get the public key from it|
|3|request times over limit|
|4|params error|
|5|the proxy is not exist|
|6|the output is illegal|
|7|system error|