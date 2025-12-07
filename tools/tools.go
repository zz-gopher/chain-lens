package tools

import "math/big"

func WeiToEther(balance *big.Int, decimals uint8) *big.Float {
	// 1. 创建一个 big.Float 类型的余额副本
	fBalance := new(big.Float).SetInt(balance)

	// 2. 计算除数 10^decimals
	base := big.NewInt(10)
	power := big.NewInt(int64(decimals)) // 这里把 uint8 转为 int64
	divisorInt := new(big.Int).Exp(base, power, nil)

	// 3. 把除数也转为 big.Float
	fDivisor := new(big.Float).SetInt(divisorInt)

	// 4. 做除法 (Balance / Divisor)
	result := new(big.Float).Quo(fBalance, fDivisor)

	return result
}
