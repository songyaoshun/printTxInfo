// package main

// import (
// 	"context"
// 	"flag"
// 	"log"
// 	"math/big"
// 	"os/exec"
// 	"sort"
// 	"strings"
// 	"sync"

// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/ethclient"

// 	"test/blobTx/printTxInfo"
// )

// func main() {
// 	// 参数
// 	rpcURL := flag.String("rpcURL", "http://127.0.0.1:9545", "以太坊 RPC URL")
// 	contractAddress := flag.String("ca", "0xcf7ed3acca5a467e9e704c703e8d87f634fb0fc9", "合约地址")
// 	startBlock := flag.Int64("start", 0, "起始区块")
// 	endBlock := flag.Int64("end", -1, "结束区块，默认最新区块")
// 	calldataPrefix := flag.String("calldata", "", "calldata的前10位")
// 	queryKeys := flag.String("query", "BlockNumber,Hash,GasUsed,4byte,func", "查询的交易或执行信息，例如Hash,GasUsed等")
// 	statusFilter := flag.Uint64("statusFilter", 2, "过滤特定Status值的交易，2表示不过滤，0表示失败交易，1表示成功交易")
// 	concurrency := flag.Int("concurrency", 10, "并行处理的区块数量")
// 	signaturesFile := flag.String("signatures", "signaturesS.json", "签名文件路径")

// 	flag.Parse()

// 	// 加载签名文件
// 	if err := printTxInfo.LoadSignatures(*signaturesFile); err != nil {
// 		log.Fatalf("Failed to load signatures file: %v", err)
// 	}

// 	// 检查是否安装了 cast 命令
// 	if _, err := exec.LookPath("cast"); err != nil {
// 		log.Fatalf("cast 命令未安装，请先安装 foundry: https://github.com/gakonst/foundry")
// 	}

// 	// 连接到以太坊客户端
// 	client, err := ethclient.Dial(*rpcURL)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
// 	}

// 	ctx := context.Background()
// 	contractAddr := common.HexToAddress(*contractAddress)

// 	// 获取最新区块
// 	var latestBlock *big.Int
// 	if *endBlock == -1 {
// 		header, err := client.HeaderByNumber(ctx, nil)
// 		if err != nil {
// 			log.Fatalf("Failed to get the latest block: %v", err)
// 		}
// 		latestBlock = header.Number
// 	} else {
// 		latestBlock = big.NewInt(*endBlock)
// 	}

// 	// 结果通道和等待组
// 	results := make(chan printTxInfo.TxInfo, 100)
// 	var wg sync.WaitGroup

// 	// 启动并行处理的 Goroutines
// 	for w := 0; w < *concurrency; w++ {
// 		wg.Add(1)
// 		go func(worker int) {
// 			defer wg.Done()
// 			for blockNumber := big.NewInt(*startBlock).Add(big.NewInt(int64(worker)), big.NewInt(int64(w))); blockNumber.Cmp(latestBlock) <= 0; blockNumber.Add(blockNumber, big.NewInt(int64(*concurrency))) {

// 				printTxInfo.ProcessBlock(ctx, client, blockNumber, contractAddr, *calldataPrefix, *statusFilter, strings.Split(*queryKeys, ","), results)
// 			}
// 		}(w)
// 	}

// 	wg.Wait()
// 	close(results)

// 	// 收集结果并排序
// 	var collectedResults []printTxInfo.TxInfo
// 	for result := range results {
// 		collectedResults = append(collectedResults, result)
// 	}

// 	sort.Slice(collectedResults, func(i, j int) bool {
// 		if collectedResults[i].BlockNumber == collectedResults[j].BlockNumber {
// 			return collectedResults[i].TransactionIndex < collectedResults[j].TransactionIndex
// 		}
// 		return collectedResults[i].BlockNumber < collectedResults[j].BlockNumber
// 	})

// 	// 打印结果
// 	for _, result := range collectedResults {
// 		printTxInfo.PrintTxInfo(result.TxData, result.ReceiptData, strings.Split(*queryKeys, ","))
// 	}
// }

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"test/blobTx/printTxInfo"
)

func main() {
	// 参数
	rpcURL := flag.String("rpcURL", "http://127.0.0.1:9545", "以太坊 RPC URL")
	contractAddress := flag.String("ca", "0xcf7ed3acca5a467e9e704c703e8d87f634fb0fc9", "合约地址")
	startBlock := flag.Int64("start", 0, "起始区块")
	endBlock := flag.Int64("end", -1, "结束区块，默认最新区块")
	calldataPrefix := flag.String("calldata", "", "calldata的前10位")
	queryKeys := flag.String("query", "BlockNumber,Hash,GasUsed,4byte,func", "查询的交易或执行信息，例如Hash,GasUsed等")
	statusFilter := flag.Uint64("statusFilter", 2, "过滤特定Status值的交易，2表示不过滤，0表示失败交易，1表示成功交易")
	concurrency := flag.Int("concurrency", 10, "并行处理的区块数量")
	signaturesFile := flag.String("signatures", "signaturesS.json", "签名文件路径")

	flag.Parse()

	// 加载签名文件
	if err := printTxInfo.LoadSignatures(*signaturesFile); err != nil {
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

	// 结果通道和等待组
	results := make(chan printTxInfo.TxInfo, 100)
	var wg sync.WaitGroup

	// 启动并行处理的 Goroutines
	for i := *startBlock; i <= latestBlock.Int64(); i += int64(*concurrency) {
		start := i
		end := i + int64(*concurrency) - 1
		if end > latestBlock.Int64() {
			end = latestBlock.Int64()
		}
		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			for blockNumber := start; blockNumber <= end; blockNumber++ {
				fmt.Printf("\r====> checking blockNum: %d\033[K", blockNumber)
				printTxInfo.ProcessBlock(ctx, client, big.NewInt(blockNumber), contractAddr, *calldataPrefix, *statusFilter, strings.Split(*queryKeys, ","), results)
			}
		}(start, end)
		time.Sleep(500 * time.Microsecond)
	}

	// 启动一个 Goroutine 来关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果并排序
	var collectedResults []printTxInfo.TxInfo
	for result := range results {
		collectedResults = append(collectedResults, result)
	}

	sort.Slice(collectedResults, func(i, j int) bool {
		if collectedResults[i].BlockNumber == collectedResults[j].BlockNumber {
			return collectedResults[i].TransactionIndex < collectedResults[j].TransactionIndex
		}
		return collectedResults[i].BlockNumber < collectedResults[j].BlockNumber
	})

	// 打印结果
	for _, result := range collectedResults {
		printTxInfo.PrintTxInfo(result.TxData, result.ReceiptData, strings.Split(*queryKeys, ","))
	}
}
