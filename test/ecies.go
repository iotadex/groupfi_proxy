package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"

	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	mathRand "math/rand"
	"os"
)

const (
	PRIVATEFILE = "src/cryptography/myECIES/privateKey.pem"
	PUBLICFILE  = "src/cryptography/myECIES/publicKey.pem"
)

// 生成指定math/rand字节长度的随机字符串
func GetRandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+?=-"
	bytes := []byte(str)
	result := []byte{}

	r := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// 生成ECC算法的公钥和私钥文件
// 根据随机字符串生成，randKey至少36位
func GenerateKey(randKey string) error {

	var err error
	var privateKey *ecdsa.PrivateKey
	var publicKey ecdsa.PublicKey
	var curve elliptic.Curve

	//一、生成私钥文件

	//根据随机字符串长度设置curve曲线
	length := len(randKey)
	//elliptic包实现了几条覆盖素数有限域的标准椭圆曲线,Curve代表一个短格式的Weierstrass椭圆曲线，其中a=-3
	if length < 224/8 {
		err = errors.New("私钥长度太短，至少为36位！")
		return err
	}

	if length >= 521/8+8 {
		//长度大于73字节，返回一个实现了P-512的曲线
		curve = elliptic.P521()
	} else if length >= 384/8+8 {
		//长度大于56字节，返回一个实现了P-384的曲线
		curve = elliptic.P384()
	} else if length >= 256/8+8 {
		//长度大于40字节，返回一个实现了P-256的曲线
		curve = elliptic.P256()
	} else if length >= 224/8+8 {
		//长度大于36字节，返回一个实现了P-224的曲线
		curve = elliptic.P224()
	}

	//GenerateKey方法生成私钥
	privateKey, err = ecdsa.GenerateKey(curve, strings.NewReader(randKey))
	if err != nil {
		return err
	}
	//通过x509标准将得到的ecc私钥序列化为ASN.1的DER编码字符串
	privateBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}
	//将私钥字符串设置到pem格式块中
	privateBlock := pem.Block{
		Type:  "ecc private key",
		Bytes: privateBytes,
	}

	//通过pem将设置好的数据进行编码，并写入磁盘文件
	privateFile, err := os.Create(PRIVATEFILE)
	if err != nil {
		return err
	}
	defer privateFile.Close()
	err = pem.Encode(privateFile, &privateBlock)
	if err != nil {
		return err
	}

	//二、生成公钥文件
	//从得到的私钥对象中将公钥信息取出
	publicKey = privateKey.PublicKey

	//通过x509标准将得到的ecc公钥序列化为ASN.1的DER编码字符串
	publicBytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}
	//将公钥字符串设置到pem格式块中
	publicBlock := pem.Block{
		Type:  "ecc public key",
		Bytes: publicBytes,
	}

	//通过pem将设置好的数据进行编码，并写入磁盘文件
	publicFile, err := os.Create(PUBLICFILE)
	if err != nil {
		return err
	}
	err = pem.Encode(publicFile, &publicBlock)
	if err != nil {
		return err
	}

	return nil
}

// 获取私钥文件里的私钥内容函数
func GetPrivateKeyByPemFile(priKeyFile string) (*ecies.PrivateKey, error) {
	//将私钥文件中的私钥读出，得到使用pem编码的字符串
	file, err := os.Open(priKeyFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	//将得到的字符串解码
	block, _ := pem.Decode(buffer)

	//使用x509将编码之后的私钥解析出来
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	//读取文件的ecdsa私钥转化成ecies私钥
	privateKeyForEcies := ecies.ImportECDSA(privateKey)

	return privateKeyForEcies, nil
}

// 获取公钥文件里的公钥内容函数
func GetPublicKeyByPemFile(pubKeyFile string) (*ecies.PublicKey, error) {
	var err error
	//从公钥文件获取钥匙字符串
	file, err := os.Open(pubKeyFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fileInfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	//将得到的字符串解码
	block, _ := pem.Decode(buffer)

	//使用x509将编码之后的公钥解析出来
	pubInner, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey := pubInner.(*ecdsa.PublicKey)

	publicKeyForEcies := ecies.ImportECDSAPublic(publicKey)

	return publicKeyForEcies, nil
}

// ECIES 公钥数据加密
func EnCryptByEcies(srcData, publicFile string) (cryptData string, err error) {
	//获取公钥数据
	publicKey, err := GetPublicKeyByPemFile(publicFile)
	if err != nil {
		return "", err
	}

	//公钥加密数据
	encryptBytes, err := ecies.Encrypt(rand.Reader, publicKey, []byte(srcData), nil, nil)
	if err != nil {
		return "", err
	}

	cryptData = hex.EncodeToString(encryptBytes)

	return
}

// ECIES 私钥数据解密
func DeCryptByEcies(cryptData, privateFile string) (srcData string, err error) {
	//获取私钥信息
	privateKey, err := GetPrivateKeyByPemFile(privateFile)
	if err != nil {
		return "", err
	}

	//私钥解密数据
	cryptBytes, _ := hex.DecodeString(cryptData)
	srcByte, err := privateKey.Decrypt(cryptBytes, nil, nil)
	if err != nil {
		fmt.Println("解密错误：", err)
		return "", err
	}
	srcData = string(srcByte)

	return
}

// aa6d9f1cb05c7285fab30eb1fa74c7839e8cb758d2d1be728ac5412b73d6b441
func Test() {
	privateKey, err := crypto.HexToECDSA("aa6d9f1cb05c7285fab30eb1fa74c7839e8cb758d2d1be728ac5412b73d6b441")
	if err != nil {
		log.Fatal(err)
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKey)
	fmt.Println(hex.EncodeToString(publicKeyBytes))
	publicKey, _ = crypto.UnmarshalPubkey(publicKeyBytes)

	//公钥加密数据
	encryptBytes, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(publicKey), crypto.FromECDSA(privateKey), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	cryptData := hexutil.Encode(encryptBytes)
	fmt.Println(cryptData)

	//私钥解密数据
	cryptBytes, _ := hexutil.Decode(cryptData)
	srcByte, err := ecies.ImportECDSA(privateKey).Decrypt(cryptBytes, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hexutil.Encode(srcByte))
}

func Test2() {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes))

	data := []byte("hello")
	hash := crypto.Keccak256Hash(data)
	fmt.Println(hash.Hex()) // 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(hexutil.Encode(signature)) // 0x789a80053e4927d0a898db8e065e948f5cf086e32f9ccaa54c1908e22ac430c62621578113ddbb62d509bf6049b8fb544ab06d36f916685a2eb8e57ffadde02301

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		log.Fatal(err)
	}

	matches := bytes.Equal(sigPublicKey, publicKeyBytes)
	fmt.Println(matches) // true

	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		log.Fatal(err)
	}

	sigPublicKeyBytes := crypto.FromECDSAPub(sigPublicKeyECDSA)
	matches = bytes.Equal(sigPublicKeyBytes, publicKeyBytes)
	fmt.Println(matches) // true

	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery id
	verified := crypto.VerifySignature(publicKeyBytes, hash.Bytes(), signatureNoRecoverID)
	fmt.Println(verified) // true
}

//0x04b8e701fdd0617634243d5cdcad1c2c157f0843f61fba5e7b603b46ee53eff875a7f8cdf85bdb315f5bb935c68c95c0b074a99b7aa6a07f49738ecdcd8b07f0e3f2dbe1a0c99a66086504fc65837538ff587cc84bc4623b98bf492bdec368976c82411861ec05bf809f8487735dfcff2aeaf0926a91b381cb0f5a379432f015825b9603c3b21597bfc09a85795b5ba139
