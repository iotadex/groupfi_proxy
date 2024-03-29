# GroupFi Proxy API

## url 
```
https://api.groupfi.ai
```

## GET /mint_nicknft, mint a nft which contains user's nick name
### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|address|string| user's address, smr address or evm address|
|name|string|letters and numbers, lowercase, 10 <= length <= 20|
|image| string, optional|image url|
### response
```json
{
    "result": true
}
```

## GET /verify, verify the addresses that belong to one group or not
### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|data|json string| hex bytes string|
```json
{
    "adds" : ["address1","address2"],
    "subs" : ["address1","address2"],
    "group" : "group contract",
    "threshold" : 1,
    "ts" : "current timestamp"
}
```
### response
```json
{
    "result": true,
    "sign": "sign hex string"
}
```

## GET /proxy/register, register a new proxy account or update the sign_account
### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|data|string| metadata that will used to mint a nft, it contains sign and timestamp|
### response
```json
{
    "result": true,
    "proxy_account": "a smr address"
}
```

## GET /proxy/account, get the proxy account for the user's evm account, using the sign_account to sign
### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|ts|int| current timestamp|
|sign|string| sign(bytes(string(ts))), using the sign_account|
### response
```json
{
    "result": true,
    "proxy_account": "a smr address"
}
```

## GET /proxy/send, send the TransactionEssence which contains a msg as metadata on shimmer L1 network
### params
|  name  |  type  |  description  |
| ------ | :---- | ---------- |
|data|string|bytes of TransactionEssence, hex format, containing a basic output with metadata feature|
|ts|int| current timestamp|
|sign|string| sign(bytes(data+ts)), using user's sign_account|
### response
```json
{
    "result": true,
    "transactionid": "transaction id"
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