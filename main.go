package main

import (
	"bufio"
	"chain-lens/modules/native"
	"encoding/json"
	"flag"
	"fmt" // 打印输出
	"log" // 日志报错
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

	fmt.Printf("📂 成功加载 %d 个钱包地址\n", len(addresses))

	// 连接节点 (Dial)
	client, err := core.NewClient(cfg.RpcURL)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Connected to EVM")

	var wg sync.WaitGroup
	startTime := time.Now()

	ethChecker, _ := native.NewChecker(client)
	//usdtChecker, _ := erc20.NewChecker(common.HexToAddress(cfg.TokenAddress), client)

	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address common.Address) {
			defer wg.Done()
			tokenBalance, err := ethChecker.BalanceOf(address)
			if err != nil {
				fmt.Printf("❌ 第 %d 个地址查询失败: %v\n", idx+1, err)
				return
			}
			fmt.Printf("✅ [%d] 地址: %s... | 余额: %s %s \n", idx+1, address.String()[:6], fmt.Sprintf("%.4f", tokenBalance.Balance), tokenBalance.Symbol)
		}(i, addr)

	}
	wg.Wait()
	fmt.Printf("🎉 任务全部完成！总耗时: %v\n", time.Since(startTime))
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
