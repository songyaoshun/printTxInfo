package printTxInfo

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
)

func extractTxData(tx *types.Transaction) map[string]interface{} {
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

	return txFields
}

func extractReceiptData(receipt *types.Receipt) map[string]interface{} {
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

	return reFields
}
