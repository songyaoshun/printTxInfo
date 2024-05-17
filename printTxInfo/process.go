package printTxInfo

import (
	"context"
	"encoding/hex"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ProcessBlock(ctx context.Context, client *ethclient.Client, blockNumber *big.Int, contractAddr common.Address, calldataPrefix string, statusFilter uint64, queryKeys []string, results chan<- TxInfo) {
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		log.Printf("Failed to get block %d: %v", blockNumber, err)
		return
	}

	for _, tx := range block.Transactions() {
		to := tx.To()
		if to != nil && *to == contractAddr {
			// 检查calldata前10位
			if calldataPrefix != "" {
				data := tx.Data()
				if len(data) < 10 || !strings.HasPrefix(hex.EncodeToString(data), calldataPrefix) {
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
			if statusFilter != 2 && receipt.Status != statusFilter {
				continue
			}

			txData := extractTxData(tx)
			receiptData := extractReceiptData(receipt)
			results <- TxInfo{
				BlockNumber:      receipt.BlockNumber.Uint64(),
				TransactionIndex: receipt.TransactionIndex,
				TxData:           txData,
				ReceiptData:      receiptData,
			}
		}
	}
}
