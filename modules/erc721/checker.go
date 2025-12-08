package erc721

import (
	"chain-lens/core"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Checker struct {
	TokenAddress common.Address
	EvmClient    *core.EvmClient
	Symbol       string
	Token        *Erc721
}

func NewChecker(tokenAddress common.Address, evmClient *core.EvmClient) (*Checker, error) {
	token, err := NewErc721(tokenAddress, evmClient.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind token %s: %w", tokenAddress.Hex(), err)
	}
	symbol, err := token.Symbol(nil)
	if err != nil {
		symbol = "UNKNOWN"
	}
	return &Checker{
		TokenAddress: tokenAddress,
		EvmClient:    evmClient,
		Symbol:       symbol,
		Token:        token,
	}, nil
}

func (c *Checker) BalanceOf(wallet common.Address) (*core.TokenBalance, error) {
	rawBalance, err := c.Token.BalanceOf(nil, wallet)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}
	return &core.TokenBalance{
		Symbol:        c.Symbol,
		Balance:       new(big.Float).SetInt(rawBalance),
		WalletAddress: wallet.String(),
	}, nil
}
