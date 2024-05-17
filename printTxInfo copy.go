package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Signature struct {
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	Signature       string          `json:"signature"`
	Inputs          json.RawMessage `json:"inputs"`
	Outputs         json.RawMessage `json:"outputs"`
	StateMutability string          `json:"stateMutability"`
}

var signaturesMap map[string]Signature

func loadSignatures(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var signatures []Signature
	if err := json.Unmarshal(data, &signatures); err != nil {
		return err
	}

	signaturesMap = make(map[string]Signature)
	for _, sig := range signatures {
		signaturesMap[sig.Signature] = sig
	}
	return nil
}

func main() {
	// 参数
	rpcURL := flag.String("rpcURL", "http://127.0.0.1:9545", "以太坊 RPC URL")
	contractAddress := flag.String("ca", "0xcf7ed3acca5a467e9e704c703e8d87f634fb0fc9", "合约地址")
	startBlock := flag.Int64("start", 0, "起始区块")
	endBlock := flag.Int64("end", -1, "结束区块，默认最新区块")
	calldataPrefix := flag.String("calldata", "", "calldata的前10位")
	queryKeys := flag.String("query", "BlockNumber,Hash,GasUsed,4byte,func", "查询的交易或执行信息，例如Hash,GasUsed等")
	statusFilter := flag.Uint64("statusFilter", 2, "过滤特定Status值的交易，2表示不过滤，0表示失败交易，1表示成功交易")

	signaturesFile := flag.String("signatures", "signaturesS.json", "签名文件路径")

	flag.Parse()

	// 加载签名文件
	if err := loadSignatures(*signaturesFile); err != nil {
		log.Fatalf("Failed to load signatures file: %v", err)
	}

	// 检查是否安装了 cast 命令
	if _, err := exec.LookPath("cast"); err != nil {
		log.Fatalf("cast 命令未安装，请先安装 foundry: https://github.com/gakonst/foundry")
	}

	// 连接到以太坊客户端
	client, err := ethclient.Dial(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	ctx := context.Background()
	contractAddr := common.HexToAddress(*contractAddress)

	// 获取最新区块
	var latestBlock *big.Int
	if *endBlock == -1 {
		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to get the latest block: %v", err)
		}
		latestBlock = header.Number
	} else {
		latestBlock = big.NewInt(*endBlock)
	}

	// 遍历区块
	for blockNumber := big.NewInt(*startBlock); blockNumber.Cmp(latestBlock) <= 0; blockNumber.Add(blockNumber, big.NewInt(1)) {
		block, err := client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			log.Printf("Failed to get block %d: %v", blockNumber, err)
			continue
		}

		for _, tx := range block.Transactions() {
			to := tx.To()
			if to != nil && *to == contractAddr {
				// 检查calldata前10位
				if *calldataPrefix != "" {
					data := tx.Data()
					if len(data) < 10 || !strings.HasPrefix(hex.EncodeToString(data), *calldataPrefix) {
						continue
					}
				}

				// 获取交易和执行信息
				receipt, err := client.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					log.Printf("Failed to get receipt for tx %s: %v", tx.Hash().Hex(), err)
					continue
				}

				// 过滤Status
				if *statusFilter != 2 && receipt.Status != *statusFilter {
					continue
				}

				queryKeysList := strings.Split(*queryKeys, ",")
				printTxInfo(tx, receipt, queryKeysList)
			}
		}
	}
}

func printTxInfo(tx *types.Transaction, receipt *types.Receipt, queryKeys []string) {
	calldata := hex.EncodeToString(tx.Data())
	functionSignature := "0x"
	if len(calldata) > 8 {
		functionSignature = "0x" + calldata[:8]
	}

	// 获取函数名称
	funcName := ""
	if functionSignature != "0x" {
		out, err := exec.Command("cast", "4", functionSignature).Output()
		if err == nil {
			funcName = strings.TrimSpace(string(out))
		} else {
			if sig, ok := signaturesMap[functionSignature]; ok {
				funcName = sig.Name
			}
		}
	}

	// 映射交易字段名
	txFields := map[string]interface{}{
		"Hash":     tx.Hash().Hex(),
		"Nonce":    tx.Nonce(),
		"GasPrice": tx.GasPrice(),
		"Gas":      tx.Gas(),
		"Value":    tx.Value(),
		"To":       tx.To().Hex(),
		"Data":     calldata,
		"4byte":    functionSignature,
		"func":     funcName,
	}

	// 映射回执字段名
	reFields := map[string]interface{}{
		"PostState":         hex.EncodeToString(receipt.PostState),
		"Status":            receipt.Status,
		"CumulativeGasUsed": receipt.CumulativeGasUsed,
		"Bloom":             fmt.Sprintf("%x", receipt.Bloom),
		"Logs":              receipt.Logs,
		"TxHash":            receipt.TxHash.Hex(),
		"ContractAddress":   receipt.ContractAddress.Hex(),
		"GasUsed":           receipt.GasUsed,
		"BlockHash":         receipt.BlockHash.Hex(),
		"BlockNumber":       receipt.BlockNumber.String(),
		"TransactionIndex":  receipt.TransactionIndex,
	}

	fmt.Println("Queried Fields:")
	for _, queryKey := range queryKeys {
		if value, ok := txFields[queryKey]; ok {
			fmt.Printf("Transaction %s: %v\n", queryKey, value)
		} else if value, ok := reFields[queryKey]; ok {
			fmt.Printf("Receipt %s: %v\n", queryKey, value)
		} else {
			fmt.Printf("Unknown query key: %s\n", queryKey)
		}
	}

	// // 打印交易的所有字段和值
	// fmt.Println("Transaction fields:")
	// for key, value := range txFields {
	// 	fmt.Printf("  %s: %v\n", key, value)
	// }

	// // 打印回执的所有字段和值
	// fmt.Println("Receipt fields:")
	// for key, value := range reFields {
	// 	fmt.Printf("  %s: %v\n", key, value)
	// }
}
