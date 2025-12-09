package erc20

import (
	"chain-lens/core"
	"chain-lens/tools"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type Checker struct {
	TokenAddress common.Address
	EvmClient    *core.EvmClient
	Token        *Token
	Decimals     uint8
	Symbol       string
}

// NewChecker initializes a Checker for the given ERC20 token.
// It binds the token contract and loads basic metadata (decimals and symbol).
// Decimals must be fetched successfully; symbol falls back to "UNKNOWN" on error.
func NewChecker(tokenAddress common.Address, evmClient *core.EvmClient) (*Checker, error) {
	token, err := NewToken(tokenAddress, evmClient.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind token %s: %w", tokenAddress.Hex(), err)
	}
	decimals, err := token.Decimals(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals for token %s: %w", tokenAddress.Hex(), err)
	}
	symbol, err := token.Symbol(nil)
	if err != nil {
		symbol = "UNKNOWN"
	}
	return &Checker{
		TokenAddress: tokenAddress,
		EvmClient:    evmClient,
		Token:        token,
		Decimals:     decimals,
		Symbol:       symbol,
	}, nil
}

func (c *Checker) BalanceOf(wallet common.Address) (*core.TokenBalance, error) {
	rawBalance, err := c.Token.BalanceOf(nil, wallet)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}
	readableBalance := tools.WeiToEther(rawBalance, c.Decimals)
	return &core.TokenBalance{
		Symbol:       c.Symbol,
		Balance:      readableBalance,
		Owner:        wallet,
		TokenAddress: c.TokenAddress,
		Success:      true,
	}, nil
}
