package gl

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/triplefi/go-logger/logger"
)

// OutLogger global logger
var OutLogger *logger.Logger

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
)

const (
	ERC_NATIVE = 0
	ERC20      = 20
	ERC721     = 721
	ERC_MANGO  = 10000
)

const (
	SCENERY_LUKSO = 10001
)

var (
	EVM_EMPTY_ADDRESS   = common.Address{}
	SOLANA_EMPTY_PUBKEY = solana.PublicKey{}
)

func CreateLogFiles() {
	var err error
	if err = os.MkdirAll("./logs", os.ModePerm); err != nil {
		log.Panic("Create dir './logs' error. " + err.Error())
	}
	if OutLogger, err = logger.New("logs/out.log", 1, 3, 0, logger.ERROR); err != nil {
		log.Panic("Create Outlogger file error. " + err.Error())
	}
}
