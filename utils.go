package main

import "github.com/ethereum/go-ethereum/core/types"

func ParseSender(tx *types.Transaction) (string, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())
	from, err := signer.Sender(tx)
	if err != nil {
		return "", err
	}
	return from.Hex(), nil
}
