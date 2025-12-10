package main

import (
	"bufio"
	"chain-lens/core"
	"chain-lens/modules/erc20"
	"chain-lens/modules/erc721"
	"chain-lens/modules/multicall"
	"chain-lens/modules/native"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	RpcURL       string `json:"rpc_url"`
	TokenAddress string `json:"token_address"`
	TokenType    string `json:"token_type"`
}

type RetryTask struct {
	Index   int
	Address common.Address
}

func main() {
	// 读取配置文件
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

	// 读取文件
	addresses, err := loadAddresses(*filePath)
	if err != nil {
		log.Fatalf("❌ 无法读取文件: %v", err)
	}

	RunApp(cfg, addresses)
}

func RunApp(cfg Config, addresses []common.Address) {
	fmt.Printf("📂 Successfully loaded %d wallet addresses\n", len(addresses))

	// 连接节点 (Dial)
	client, err := core.NewClient(cfg.RpcURL)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Connected to EVM")
	startTime := time.Now()
	multicallChecker, _ := multicall.NewMultiChecker(client.Client)
	tokenType, err := ParseTokenType(cfg.TokenType)
	if err != nil {
		log.Fatal(err)
	}
	tokenBalances, err := multicallChecker.CheckToken(tokenType, common.HexToAddress(cfg.TokenAddress), addresses)

	// 准备重试任务列表
	var retryTasks []RetryTask

	if err != nil {
		// --- 情况 A: Multicall 整体失败 (比如 RPC 不支持，或者合约报错) ---
		fmt.Printf("⚠️ Multicall 整体失败: %v，切换全量并发查询模式...\n", err)
		tokenBalances = make([]core.TokenBalance, len(addresses))
		// 所有地址都要重试
		for i, addr := range addresses {
			retryTasks = append(retryTasks, RetryTask{Index: i, Address: addr})
		}
	} else {
		// --- 情况 B: Multicall 成功，但可能有部分个例失败 ---
		for i, tb := range tokenBalances {
			if !tb.Success {
				retryTasks = append(retryTasks, RetryTask{Index: i, Address: tb.Owner})
			}
		}
	}
	// 执行并发补救 (如果有失败任务)
	if len(retryTasks) > 0 {
		fmt.Printf("🔄 开始并发修补 %d 个失败任务...\n", len(retryTasks))

		var wg sync.WaitGroup
		var mu sync.Mutex // 关键：保护 tokenBalances 的写锁

		// 信号量：限制并发数 (比如限制 20 个并发)，防止把 RPC 节点打挂
		sem := make(chan struct{}, 20)

		// 初始化单次查询器 (Fallback Checker)
		singleChecker := NewTokenChecker(cfg, client)

		for _, task := range retryTasks {
			wg.Add(1)
			sem <- struct{}{} // 拿令牌

			go func(t RetryTask) {
				defer wg.Done()
				defer func() { <-sem }() // 还令牌

				// 执行单次查询
				singleResult, err := singleChecker.BalanceOf(t.Address)

				// 加锁回写数据
				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					// 彻底失败：记录错误
					fmt.Printf("❌ 重试仍失败 [%d] %s: %v\n", t.Index, t.Address.Hex(), err)
					// 确保结果数组里对应的位置有标记
					tokenBalances[t.Index].Owner = t.Address
					tokenBalances[t.Index].Success = false
				} else {
					// 🎉 挽救成功：更新原本的数据
					fmt.Printf("✅ 修补成功 [%d] %s\n", t.Index, t.Address.Hex())
					// 这里要把 singleResult 转换成 TokenBalance 格式填回去
					tokenBalances[t.Index] = core.TokenBalance{
						TokenAddress: common.HexToAddress(cfg.TokenAddress),
						Owner:        t.Address,
						Balance:      singleResult.Balance, // 假设 singleResult 结构
						Symbol:       singleResult.Symbol,
						Success:      true, // 标记为成功
					}
				}
			}(task)
		}
		wg.Wait()
	}
	// 最终统计
	//idexList := make([]int, 0, 100)
	totalBalance := new(big.Float)
	successCount := 0
	for idx, tb := range tokenBalances {
		if tb.Success {
			successCount++
			// 🔒 安全检查：防止 tb.Balance 为 nil 导致 panic
			if tb.Balance != nil {
				// 累加逻辑: totalBalance = totalBalance + tb.Balance
				totalBalance.Add(totalBalance, tb.Balance)
			}
			//if tb.Balance.Cmp(big.NewFloat(1)) >= 0 {
			//	// 大于等于 1
			//	idexList = append(idexList, idx+1)
			//}
			// 这里可以打印最终结果
			fmt.Printf("✅ [%d] Address: %s... | Balance: %s %s \n", idx+1, tb.Owner.String()[:6], fmt.Sprintf("%.4f", tb.Balance), tb.Symbol)
		}
	}
	fmt.Printf("\n--------------------------------------------------\n")
	fmt.Printf("📊 Summary Report\n")
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("✅ Success Rate : %d / %d\n", successCount, len(addresses))

	// 格式化输出:
	// %.4f 表示保留 4 位小数
	// big.Float 实现了 fmt.Formatter 接口，可以直接这样打印
	fmt.Printf("💰 Total Balance: %.4f %s\n ", totalBalance, tokenBalances[0].Symbol)
	fmt.Printf("🎉 All tasks completed! Success: %d/%d | Time: %v\n", successCount, len(addresses), time.Since(startTime))
	fmt.Printf("--------------------------------------------------\n")

	//for _, v := range idexList {
	//	fmt.Printf("%d ", v)
	//}
}

// NewTokenChecker creates a token checker.
// Uses cfg.TokenType if set; otherwise auto-detects ERC20 → ERC721 → native.
// Program exits if all attempts fail.
func NewTokenChecker(cfg Config, evmClient *core.EvmClient) core.AssetChecker {
	tokenAddr := common.HexToAddress(cfg.TokenAddress)
	var checker core.AssetChecker
	var err error
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

func ParseTokenType(s string) (multicall.TokenType, error) {
	if s == "" {
		return 0, fmt.Errorf("❌ Configuration Error: 'token_type' in config file cannot be empty. Valid values are: native, erc20, erc721")
	}
	switch strings.ToLower(s) {
	case "native":
		return multicall.TokenTypeNative, nil
	case "erc20":
		return multicall.TokenTypeERC20, nil
	case "erc721":
		return multicall.TokenTypeERC721, nil
	default:
		return 0, fmt.Errorf("❌ Configuration Error: Invalid token_type. Valid values are: native, erc20, erc721")
	}
}

// ChunkSlice 把一个大的钱包地址列表，切分成多个小批次
// 例如：输入 5 个地址，batchSize 是 2 -> 输出 [[1,2], [3,4], [5]]
func ChunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		// 防止越界
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
