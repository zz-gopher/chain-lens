package main

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func randomAddress() common.Address {
	b := make([]byte, 20)
	rand.Read(b)
	return common.HexToAddress("0x" + hex.EncodeToString(b))
}

func TestRunApp_ThousandWallets(t *testing.T) {
	// 生成 1000 个假钱包
	var addrs []common.Address
	for i := 0; i < 10000; i++ {
		addrs = append(addrs, randomAddress())
	}

	cfg := Config{
		RpcURL:       "https://rpc.soneium.org", // mock
		TokenAddress: "0x102d758f688a4c1c5a80b116bd945d4455460282",
		TokenType:    "erc20",
	}

	RunApp(cfg, addrs)

	t.Log("✓ 1000 地址测试通过!")
}
