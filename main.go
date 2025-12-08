package main

import (
	"bufio"
	"chain-lens/modules/erc20"
	"chain-lens/modules/erc721"
	"chain-lens/modules/native"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"chain-lens/core"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	RpcURL       string `json:"rpc_url"`
	TokenAddress string `json:"token_address"`
	TokenType    string `json:"token_type"`
}

func main() {
	// 1. 读取配置文件
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("请确保目录下有 config.json 文件")
	}

	var cfg Config
	if err := json.Unmarshal(configFile, &cfg); err != nil {
		log.Fatal("配置解析失败，请检查 json 格式")
	}

	filePath := flag.String("file", "wallets.txt", "包含钱包地址的文件路径 (每行一个)")
	flag.Parse()

	// 2. 读取文件
	addresses, err := loadAddresses(*filePath)
	if err != nil {
		log.Fatalf("❌ 无法读取文件: %v", err)
	}

	fmt.Printf("📂 Successfully loaded %d wallet addresses\n", len(addresses))

	// 连接节点 (Dial)
	client, err := core.NewClient(cfg.RpcURL)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Connected to EVM")

	var wg sync.WaitGroup
	startTime := time.Now()

	//ethChecker, _ := native.NewChecker(client)
	//usdtChecker, _ := erc20.NewChecker(common.HexToAddress(cfg.TokenAddress), client)
	//erc721Checker, _ := erc721.NewChecker(common.HexToAddress(cfg.TokenAddress), client)

	checker := NewTokenChecker(cfg, client)

	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address common.Address) {
			defer wg.Done()
			tokenBalance, err := checker.BalanceOf(address)
			if err != nil {
				fmt.Printf("❌ 第 %d 个地址查询失败: %v\n", idx+1, err)
				return
			}
			fmt.Printf("✅ [%d] Address: %s... | Balance: %s %s \n", idx+1, address.String()[:6], fmt.Sprintf("%.4f", tokenBalance.Balance), tokenBalance.Symbol)
		}(i, addr)

	}
	wg.Wait()
	fmt.Printf("🎉 All tasks completed! Total elapsed time: %v\n", time.Since(startTime))
}

// NewTokenChecker creates a token checker.
// Uses cfg.TokenType if set; otherwise auto-detects ERC20 → ERC721 → native.
// Program exits if all attempts fail.
func NewTokenChecker(cfg Config, evmClient *core.EvmClient) core.AssetChecker {
	tokenAddr := common.HexToAddress(cfg.TokenAddress)
	var checker core.AssetChecker
	var err error
	// 用户指定类型
	if cfg.TokenType != "" {
		switch cfg.TokenType {
		case "native":
			checker, err = native.NewChecker(evmClient)
		case "erc20":
			checker, err = erc20.NewChecker(tokenAddr, evmClient)
		case "erc721":
			checker, err = erc721.NewChecker(tokenAddr, evmClient)
		default:
			log.Fatalf("❌ Failed to create checker for token: %s", cfg.TokenAddress)
		}
		if err != nil {
			log.Fatalf("❌ Failed to create checker for token: %s", cfg.TokenAddress)
		}
	} else {
		// 自动识别 ERC20 → ERC721 → native
		checker, err = erc20.NewChecker(tokenAddr, evmClient)
		if err == nil {
			fmt.Println("🔹 Auto-detect ERC20 token")
			return checker
		}

		checker, err = erc721.NewChecker(tokenAddr, evmClient)
		if err == nil {
			fmt.Println("🔹 Auto-detect ERC721 token")
			return checker
		}

		checker, err = native.NewChecker(evmClient)
		if err == nil {
			fmt.Println("🔹 Auto-detect native token")
			return checker
		}

		// 所有方式都失败
		log.Fatalf("❌ Failed to create checker for token: %s", cfg.TokenAddress)
	}
	return nil
}

func loadAddresses(path string) ([]common.Address, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var addresses []common.Address
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// 校验是否为合法地址
		if !common.IsHexAddress(line) {
			log.Printf("⚠️ 跳过无效地址: %s", line)
			continue
		}
		addresses = append(addresses, common.HexToAddress(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return addresses, nil
}
