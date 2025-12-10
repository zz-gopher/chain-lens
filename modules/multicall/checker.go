package multicall

import (
	"chain-lens/core"
	"chain-lens/modules/erc20"
	"chain-lens/modules/erc721"
	"chain-lens/tools"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TokenType int

const ContractAddress = "0xcA11bde05977b3631167028862bE2a173976CA11"

const (
	TokenTypeERC20 TokenType = iota
	TokenTypeERC721
	TokenTypeNative
)

type MultiChecker struct {
	Client        *ethclient.Client
	Multicall     *Multicall
	MulticallAddr common.Address
}

type callItem struct {
	Token    common.Address
	Owner    common.Address
	Type     TokenType
	CallData []byte
	AbiName  string
}

func NewMultiChecker(client *ethclient.Client) (*MultiChecker, error) {
	multicallAddr := common.HexToAddress(ContractAddress)
	// 绑定multicall合约
	multi, err := NewMulticall(multicallAddr, client)
	if err != nil {
		return nil, err
	}
	return &MultiChecker{
		Client:        client,
		Multicall:     multi,
		MulticallAddr: multicallAddr,
	}, nil
}

func (m *MultiChecker) CheckToken(tType TokenType, tokenAddr common.Address, owners []common.Address) ([]core.TokenBalance, error) {
	var callList []callItem
	var decimals uint8
	var symbol string
	// 准备 Multicall3 的 ABI，用于 Native 代币打包
	mcAbi, err := MulticallMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get multicall abi: %w", err)
	}
	switch tType {
	case TokenTypeERC20:
		items, err := setCallList(erc20.TokenMetaData, owners, tokenAddr, tType)
		if err != nil {
			return nil, err
		}
		callList = append(callList, items...)
		// 绑定erc20合约
		token, err := erc20.NewToken(tokenAddr, m.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to bind token %s: %w", tokenAddr.Hex(), err)
		}
		// 查询代币精度
		decimals, err = token.Decimals(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get decimals for token %s: %w", tokenAddr.Hex(), err)
		}
		symbol, err = token.Symbol(nil)
		if err != nil {
			symbol = "UNKNOWN"
		}
	case TokenTypeERC721:
		decimals = 1
		symbol = "NFT"
		items, err := setCallList(erc721.Erc721MetaData, owners, tokenAddr, tType)
		if err != nil {
			return nil, err
		}
		callList = append(callList, items...)
	case TokenTypeNative:
		decimals = 18
		symbol = "ETH"
		for _, owner := range owners {
			callData, err := mcAbi.Pack("getEthBalance", owner)
			if err != nil {
				return nil, fmt.Errorf("pack getEthBalance failed: %w", err)
			}

			callList = append(callList, callItem{
				Token:    tokenAddr, // 这里的 Token 只是标识，Target 在下面会换成 MulticallAddr
				Owner:    owner,
				Type:     TokenTypeNative,
				CallData: callData,
				AbiName:  "getEthBalance",
			})
		}
	default:
		return nil, errors.New("unknown token type")
	}

	var mcCalls []Multicall3Call3
	for _, c := range callList {
		target := c.Token
		// 关键修正：如果查原生代币，Target 必须是 Multicall 合约地址本身
		if c.Type == TokenTypeNative {
			target = m.MulticallAddr
		}

		mcCalls = append(mcCalls, Multicall3Call3{
			Target:       target,
			CallData:     c.CallData,
			AllowFailure: true, // 允许部分失败
		})
	}
	// 执行multicall3的Aggregate3,把多个合约调用封装（Pack）成一个大调用，一次性发给区块链执行
	resp, err := m.Multicall.Aggregate3(nil, mcCalls)
	if err != nil {
		return nil, fmt.Errorf("multicall aggregate3 failed: %w", err)
	}
	var balances []core.TokenBalance
	erc20Abi, _ := erc20.TokenMetaData.GetAbi()
	erc721Abi, _ := erc721.Erc721MetaData.GetAbi()
	for i, res := range resp {
		req := callList[i]
		tb := core.TokenBalance{
			TokenAddress: req.Token,
			Owner:        req.Owner,
			Balance:      big.NewFloat(0), // 默认为 0
			Symbol:       symbol,
		}

		// 检查是否调用成功
		if !res.Success {
			balances = append(balances, tb)
			tb.Success = false
			continue
		}
		// 检查返回数据是否为空
		if len(res.ReturnData) == 0 {
			// ERC20可能虽然没查到数据，但我们可以“认为”它余额是 0
			// 因为一个不存在的合约，你当然没有它的币
			balances = append(balances, tb)
			continue
		}
		// 视为查询成功
		tb.Success = true
		// 根据类型解码
		var decodeErr error
		var rawBalance *big.Int
		switch req.Type {
		case TokenTypeERC20:
			// ERC20 解码
			rawBalance, decodeErr = decodeUint256(erc20Abi, "balanceOf", res.ReturnData)
			tb.Balance = tools.WeiToEther(rawBalance, decimals)
		case TokenTypeERC721:
			// ERC721 解码
			rawBalance, decodeErr = decodeUint256(erc721Abi, "balanceOf", res.ReturnData)
			tb.Balance = new(big.Float).SetInt(rawBalance)
		case TokenTypeNative:
			// Native 解码 (getEthBalance 返回 uint256)
			// 直接由 bytes 转 bigInt 即可，或者用 ABI unpack 也可以
			rawBalance := new(big.Int).SetBytes(res.ReturnData)
			tb.Balance = tools.WeiToEther(rawBalance, decimals)
		}

		if decodeErr != nil {
			return nil, decodeErr
		}

		balances = append(balances, tb)
	}

	return balances, nil
}

// 辅助函数：修改 setCallList 以接受 type
func setCallList(metaData *bind.MetaData, owners []common.Address, tokenAddr common.Address, tType TokenType) ([]callItem, error) {
	var callList []callItem
	parsed, err := metaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("parse abi: %w", err)
	}
	for _, owner := range owners {
		// 把函数+参数->编码成EVM需要的calldata
		data, err := parsed.Pack("balanceOf", owner)
		if err != nil {
			return nil, fmt.Errorf("pack balanceOf: %w", err)
		}
		callList = append(callList, callItem{
			Token:    tokenAddr,
			Owner:    owner,
			Type:     tType, // 使用传入的 type
			CallData: data,
			AbiName:  "balanceOf",
		})
	}
	return callList, nil
}

// 辅助函数：通用解码 Uint256
func decodeUint256(parsedAbi *abi.ABI, method string, data []byte) (*big.Int, error) {
	// 1. 解码二进制数据
	unpacked, err := parsedAbi.Unpack(method, data)
	if err != nil {
		return nil, err
	}

	// 2. 校验数据完整性
	if len(unpacked) == 0 {
		return nil, errors.New("no data unpacked")
	}

	// 3. 类型断言 (Type Assertion)
	balance, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("result is not *big.Int, type is %T", unpacked[0])
	}
	return balance, nil
}
