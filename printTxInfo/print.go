package printTxInfo

import "fmt"

func PrintTxInfo(txData, receiptData map[string]interface{}, queryKeys []string) {
	fmt.Println("\n===> Queried Fields:")
	for _, queryKey := range queryKeys {
		if value, ok := txData[queryKey]; ok {
			fmt.Printf("Transaction %s: %v\n", queryKey, value)
		} else if value, ok := receiptData[queryKey]; ok {
			fmt.Printf("Receipt %s: %v\n", queryKey, value)
		} else {
			fmt.Printf("Unknown query key: %s\n", queryKey)
		}
	}

	// // 打印交易的所有字段和值
	// fmt.Println("Transaction fields:")
	// for key, value := range txData {
	// 	fmt.Printf("  %s: %v\n", key, value)
	// }

	// // 打印回执的所有字段和值
	// fmt.Println("Receipt fields:")
	// for key, value := range receiptData {
	// 	fmt.Printf("  %s: %v\n", key, value)
	// }
}
