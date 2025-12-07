package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenBalance 统一的返回结果结构
type TokenBalance struct {
	Symbol        string     // 代币符号
	Balance       *big.Float // 余额
	WalletAddress string     // 钱包地址
}

// AssetChecker 定义通用的查余额接口
type AssetChecker interface {
	BalanceOf(address common.Address) (*TokenBalance, error)
}
