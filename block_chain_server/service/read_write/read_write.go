// Package readwrite: task1: read write block chain
package read_write

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 编写 Go 代码，使用 ethclient 连接到 Sepolia 测试网络。
// 实现查询指定区块号的区块信息，包括区块的哈希、时间戳、交易数量等。
// 输出查询结果到控制台。

func mustRPCURL() string {
	url := os.Getenv("SEPOLIA_RPC_URL")
	if url == "" {
		log.Fatal("SEPOLIA_RPC_URL not set")
	}
	return url
}

func diaClient(ctx context.Context) *ethclient.Client {
	client, err := ethclient.DialContext(ctx, mustRPCURL())
	if err != nil {
		log.Fatalf("dial sepolia failed: %v", err)
	}
	return client
}

func QueryBlockInfo(blockNum int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := diaClient(ctx)
	defer client.Close()

	var number *big.Int
	// if blockNum = nil , get the latest block
	if blockNum > 0 {
		number = big.NewInt(blockNum)
	}

	block, err := client.BlockByNumber(ctx, number)
	if err != nil {
		return fmt.Errorf("get block failed: %w", err)
	}

	ts := time.Unix(int64(block.Time()), 0)
	fmt.Printf("Block Number: %v\n", block.Number())
	fmt.Printf("Block Hash: %s\n", block.Hash().Hex())
	fmt.Printf("Timestamp: %s\n", ts.Format(time.RFC3339))
	fmt.Printf("Tx Count: %d\n", len(block.Transactions()))
	return nil
}

func SendETH(toHex string, amountWei *big.Int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := diaClient(ctx)
	defer client.Close()

	// get private key and address
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return "", fmt.Errorf("PRIVATE_KEY not set")
	}

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("bad private key: %w", err)
	}
	// fromAdrr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("cannot assert type: publicKey is not *ecdsa.PublicKey")
	}
	fromAdrr := crypto.PubkeyToAddress(*publicKeyECDSA)

	// get chain id -- sepolia
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return "", fmt.Errorf("get chain id: %w", err)
	}

	// get fee's tip
	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return "", fmt.Errorf("suggest tip: %w", err)
	}

	// get header number
	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}

	// get base fee
	baseFee := head.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(0)
	}

	// calculate fee cap cap : basefee * 2 + tip
	feeCap := new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(2)), tip)

	// get nonce
	nonce, err := client.PendingNonceAt(ctx, fromAdrr)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	to := common.HexToAddress(toHex)

	// estimate gas limit
	callMsg := ethereum.CallMsg{
		From:  fromAdrr,
		To:    &to,
		Value: amountWei,
		Data:  nil,
	}
	gasLimit, err := client.EstimateGas(ctx, callMsg)
	if err != nil {
		gasLimit = 21000
	}

	// make eip-1559 transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     amountWei,
		Data:      nil,
	})

	// sign and send transaction
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKeyECDSA)
	if err != nil {
		return "", fmt.Errorf("sign tx : %w", err)
	}
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("send tx : %w", err)
	}
	return signedTx.Hash().Hex(), nil
}
