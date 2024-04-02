package tools

type Cryptor interface {
	GetEncryptString(source string, seeds [4]uint64) []byte
	GetDecryptString(encrypt string, seeds [4]uint64) []byte
	SignDataByECDSA(data string, seeds [4]uint64) ([]byte, error)
	Input()              // input the password to create seeds
	ReadRand() [4]uint64 // read the seeds to memory
}

var Aes Cryptor
