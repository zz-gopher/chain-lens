package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenBalance 统一的返回结果结构
type TokenBalance struct {
	Symbol       string         // 代币符号
	TokenAddress common.Address // 代币合约地址
	Balance      *big.Float     // 余额
	Owner        common.Address // 钱包地址
	Success      bool           // 是否查询成功
}

// AssetChecker 定义通用的查余额接口
type AssetChecker interface {
	BalanceOf(address common.Address) (*TokenBalance, error)
}
