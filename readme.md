# GroupFi Proxy API

## url 
```
https://api.groupfi.ai
```

## public apis
### GET /hornet, get a hornet node url
#### response
```json
{
    "result": true,
    "id":     1,
    "rpc": "https://example.com"
}
```

### GET /chains, get all the evm chains that grougfi supported
#### response
```json
{
    "1": {
        "chainid": 1,
        "name": "ethereum",
        "Symbol": "ETH",
        "decimal": 18,
        "contract": "0xD735F8c06c8b802E989512F85E696b11733d7734",
        "pic_uri": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets"
    },
    "148": {
        "chainid": 148,
        "name": "shimmer",
        "Symbol": "SMR",
        "decimal": 18,
        "contract": "0xAEaDcd57E4389678537d82891f095BBbE0ab9610",
        "pic_uri": ""
    }
}
```

### GET /rpc?chainid={42}, get the rpc uri by chain id. Lukso is the only supported chainid now!
#### response
```json
{
    "result": true,
    "rpc": "https://example.com"
}
```

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

### GET /group/checkname?n={name1}, check the is valid or not.
#### response
```json
{
    "result" : true,
}
```
or
```json
{
    "result" : false,
    "err-code" : 4
}
```

### POST /group/dids, get dids for all the addresses
#### params
```json
{
    "addresses" : ["address1","address2"],
    "updates" : [false, false]
}
```
the length of `addresses` and `updates` must be same.

#### response
```json
{
    "result": true,
    "dids": {
        "address1" : {
            "148" : {
                "name" : "John",
                "image_url" : ""
            },
            "42" : {
                "name": "Hello",
                "image_url": ""
            }
        },
        "address2" : {
            "148" : {
                "name" : "Rodor",
                "image_url" : ""
            },
            "42" : {
                "name": "Wulfgar",
                "image_url": ""
            }
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
erc is 20 or 721 or 0, and 0 represents native token.
the contract is `11111111111111111111111111111111` if solana chain
the addresses are the solana's main account always
#### response
```json
{
    "result": true,
    "indexes": [],
}
```

### POST /group/filter/v2, filter the addresses
#### params
```json
{
    "addresses" : ["address1","address2"],
    "chains" : [
        {
            "chain": 148,
            "contract" : "group contract address",
            "threshold" : "1000000000",
            "erc" : 721
        }
    ],
    "ts" : 1712647238
}
```
erc is 20 or 721 or 0, and 0 represents native token.
the contract is `11111111111111111111111111111111` if solana chain
the addresses are the solana's main account always
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

### GET /faucet?chainid={1}&token={2}&to={3}&amount={4}, send test token to user
#### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|chainid|int64| evm network chain id|
|token|string|erc20 token's contract address|
|to| string|user's address, every address can get the test token per day|
|amount|string|token's amount|

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
    "proxy_account": "smr address",
    "outputids": ["0x07aa975f6c267c33d77cd9900e2f4e3dab52608b118ba38cb014a8a23cb764bb0200"],
    "outputs": [{"basicOutput"}]
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



### POST /proxy/send?hornet={https://production.api.groupfi.ai, optional}, send the TransactionEssence which contains a msg as metadata on shimmer L1 network
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

### POST /proxy/send/asyn?hornet={production, optional}, send the TransactionEssence which contains a msg as metadata on shimmer L1 network, asynchronization
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

## Groups Supported

| Group Name |   Erc  |   Chain     | Description |
| ---------- | ------ | ----------- | ----------- |
| erc20      |   20   | all evm     | erc20 token |
| erc721     |   721  | all evm     | erc721 NFT  |
| spl        |   20   | solana      | spl token   |
| native     |   0    | evm,solana  | native token|
| mango      | 10000  | solana      | mango market|
| erc404     | 404    | all evm     | erc404 token|
| erc72100   | 72100  | partial evm | token_uri   |
| erc115500  | 115500 | partial evm | token_uri   |

## Error Response
### Response format
```json
{
    "result"  : false,
    "err-code": 0,
    "err-msg" : "system error"
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