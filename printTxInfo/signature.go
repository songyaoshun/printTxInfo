package printTxInfo

import (
	"encoding/json"
	"io/ioutil"
)

type Signature struct {
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	Signature       string          `json:"signature"`
	Inputs          json.RawMessage `json:"inputs"`
	Outputs         json.RawMessage `json:"outputs"`
	StateMutability string          `json:"stateMutability"`
}

type TxInfo struct {
	BlockNumber      uint64
	TransactionIndex uint
	TxData           map[string]interface{}
	ReceiptData      map[string]interface{}
}

var signaturesMap map[string]Signature

func LoadSignatures(filePath string) error {
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
