package native

import (
	"chain-lens/core"
	"chain-lens/tools"
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Checker struct {
	EvmClient *ethclient.Client
}

func NewChecker(evmClient *core.EvmClient) (*Checker, error) {
	return &Checker{
		EvmClient: evmClient.Client,
	}, nil
}

// BalanceOf CheckBalance 查ETH余额的工具函数
func (c *Checker) BalanceOf(address common.Address) (*core.TokenBalance, error) {
	weiBalance, err := c.EvmClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	ethValue := tools.WeiToEther(weiBalance, 18)
	return &core.TokenBalance{
		Symbol:        "ETH",
		Balance:       ethValue,
		WalletAddress: address.String(),
	}, nil
}
