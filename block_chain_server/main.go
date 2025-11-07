package main

import (
	"block_chain_server/service/deplpy_contracts"
	readwrite "block_chain_server/service/read_write"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
)

/*
用法：
 1. 查询最新区块： go run .
 2. 查询指定区块： go run . 123456
 3. 发送转账：     go run . send 0xReceiverAddress 10000000000000000
    (以上金额单位为 wei，例子代表 0.01 ETH)
*/
func main() {
	// choose action by args:
	//   - "deploy": deploy and interact with Counter contract
	//   - otherwise: read/write helpers (latest block, specific block, send)

	// readWriteBlockChain()
	deplpy_contracts.DeployCountContract()
}

func readWriteBlockChain() {
	if len(os.Args) > 1 && os.Args[1] == "send" {
		if len(os.Args) < 4 {
			log.Fatal("usage go run . send <to> <amountwei>")
		}
		to := os.Args[2]
		n, ok := new(big.Int).SetString(os.Args[3], 10)
		if !ok {
			log.Fatal("invalid amountwei")
		}
		hash, err := readwrite.SendETH(to, n)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Tx Hash:", hash)
		return
	}
	var num int64 = -1
	if len(os.Args) > 1 {
		n, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil {
			num = n
		}
	}
	if err := readwrite.QueryBlockInfo(num); err != nil {
		log.Fatal(err)
	}
}
