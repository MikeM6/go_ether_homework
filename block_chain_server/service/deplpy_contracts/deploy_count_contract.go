package deplpy_contracts

import (
	"block_chain_server/contracts/counter"
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func DeployCountContract() {
	rpcURL := os.Getenv("SEPOLIA_RPC_URL")
	pkHey := os.Getenv("PRIVATE_KEY")

	if rpcURL == "" || pkHey == "" {
		log.Fatal("set SEPOLIA_RPC_URL and PRIVATE_KEY")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(pkHey, "0x"))
	if err != nil {
		log.Fatal(err)
	}

	// publicKey := privateKey.Public()
	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	log.Fatal("get public key ecdsa error")
	// }
	// fromAdrr := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID := big.NewInt(11155111)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}
	auth.Value = big.NewInt(0)
	auth.GasLimit = 0

	// use abigen deplpy contract
	addr, tx, instance, err := counter.DeployCounter(auth, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deployed address:", addr.Hex(), "tx", tx.Hash().Hex())

	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal(err)
	}
	// read count contract value
	val, err := instance.Value(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("value before:", val)
	// execute increment
	tx2, err := instance.Increment(auth, big.NewInt(3))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("increment tx:", tx2.Hash().Hex())

	_, err = bind.WaitMined(context.Background(), client, tx2)
	if err != nil {
		log.Fatal(err)
	}
	val2, err := instance.Value(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("value after:", val2)

}
