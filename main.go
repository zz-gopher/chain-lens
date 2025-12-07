package main

import (
	"bufio"
	"context" // 上下文控制
	"encoding/json"
	"flag"
	"fmt"      // 打印输出
	"log"      // 日志报错
	"math/big" // 大数计算
	"os"
	"strings"
	"sync"
	"time"

	"chain-lens/token"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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

	// 1. 连接节点 (Dial)
	client, err := ethclient.Dial(cfg.RpcURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Connected to Ethereum")

	var wg sync.WaitGroup
	startTime := time.Now()

	hexAddress := common.HexToAddress(cfg.TokenAddress)
	instance, err := token.NewToken(hexAddress, client)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
	}
	decimals, err := instance.Decimals(nil)
	if err != nil {
		log.Fatal(err)
	}

	symbol, err := instance.Symbol(nil)
	if err != nil {
		symbol = "UNKNOWN"
	}

	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address common.Address) {
			defer wg.Done()
			balance, err := checkErc20Balance(instance, address.String(), decimals)
			if err != nil {
				fmt.Printf("❌ 第 %d 个地址查询失败: %v\n", idx+1, err)
				return
			}
			fmt.Printf("✅ [%d] 地址: %s... | 余额: %s %s \n", idx+1, address.String()[:6], fmt.Sprintf("%.4f", balance), symbol)
		}(i, addr)

	}
	wg.Wait()
	fmt.Printf("🎉 任务全部完成！总耗时: %v\n", time.Since(startTime))
}

// 查ETH余额的工具函数
func checkBalance(client *ethclient.Client, address string) (*big.Float, error) {
	account := common.HexToAddress(address)
	weiBalance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, err
	}
	ethValue := weiToEther(weiBalance, 18)
	return ethValue, nil
}

// 查 Erc20 token的工具函数
func checkErc20Balance(instance *token.Token, address string, decimals uint8) (*big.Float, error) {
	account := common.HexToAddress(address)
	rawBalance, err := instance.BalanceOf(nil, account)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}
	// 6. 计算最终金额
	readableBalance := weiToEther(rawBalance, decimals)

	// 7. 打包返回
	return readableBalance, nil

}

func weiToEther(balance *big.Int, decimals uint8) *big.Float {
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
