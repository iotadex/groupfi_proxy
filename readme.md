# GroupFi Proxy API

## url 
```
https://api.groupfi.ai
```

## GET /mint_nicknft, mint a nft which contains user's nick name
### params
|  name  |  type  |  descript  |
| ------ | :---- | ---------- |
|address|string| user's address, smr address or evm address|
|name|string|letters and numbers, lowercase, 10 <= length <= 20|
|image| string, optional|image url|
### response
```json
{
    "result":   true,
	"block_id": "id of the block which mints the name fnt",
}
```

## GET /proxy/account, get the proxy account for the user's evm account, using the sign_account to sign
### params
|  name  |  type  |  descript  |
| ------ | :---- | ---------- |
|ts|int| current timestamp|
|sign|string| sign(bytes(string(ts))) using the sign account|
### response
```json
{
    "result":   true,
	"proxy_account": "a smr address",
}
```

## GET /proxy/register, register a new proxy account or update the sign account
### params
|  name  |  type  |  descript  |
| ------ | :---- | ---------- |
|chain|string|the network symbol, eth, bsc, shimmer, iota, etc|
|data|string|the sign_account's address, a evm address|
|ts|int| current timestamp|
|sign|string| sign(bytes(chain+data+ts)) using user's evm main account|
### response
```json
{
    "result":   true,
	"proxy_account": "a smr address",
}
```

## GET /proxy/sign_account, sign the TransactionEssence which store the map fo evm account<->sign_account to shimmer L1 network
### params
|  name  |  type  |  descript  |
| ------ | :---- | ---------- |
|data|string|bytes of TransactionEssence, hex format, the tx contains a nft|
|ts|int| current timestamp|
|sign|string| sign(bytes(data+ts)) using user's evm main account|
### response
```json
{
    "result":   true,
	"proxy_account": "a smr address",
}
```