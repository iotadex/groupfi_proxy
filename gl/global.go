package gl

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
)

type ERRORCODE int

const (
	TIMEOUT_ERROR      ERRORCODE = iota + 1 // the signed ts is time out, 10 minutes
	SIGN_ERROR                              // sign error, can not get the public key from it
	REQUEST_LIMIT                           // request times over limit
	PARAMS_ERROR                            // params error
	PROXY_NOT_EXIST                         // the proxy is not exist
	MSG_OUTPUT_ILLEGAL                      // the output is illegal
	SYSTEM_ERROR                            // system error
)

const (
	SOLANA_CHAINID = 518
	LUKSO_CHAINID  = 42
)

const (
	ERC_NATIVE = 0
	ERC20      = 20
	ERC721     = 721
	ERC_MANGO  = 10000
)

var (
	EVM_EMPTY_ADDRESS   = common.Address{}
	SOLANA_EMPTY_PUBKEY = solana.PublicKey{}
	LUKSO_UP_HELP       = common.HexToAddress("0x0A86EcF432Bb889Fc000804ecF04b4A96017fC78")
)
