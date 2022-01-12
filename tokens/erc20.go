package tokens

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/weijun-sh/reconciliationSystem-bridge/common"
)

var getTokenCount int = 3

var erc20CodeParts = map[string][]byte{
	"name":        common.FromHex("0x06fdde03"),
	"symbol":      common.FromHex("0x95d89b41"),
	"decimal":     common.FromHex("0x313ce567"),
	"balanceOf":   common.FromHex("0x70a08231"),
	"totalsupply": common.FromHex("0x18160ddd"),
}

var Decimal map[string]*big.Float = make(map[string]*big.Float)

// GetErc20Balance get erc20 balacne of address
func getErc20Balance(client *ethclient.Client, contract, address string) (*big.Float, error) {
	//fmt.Printf("getErc20Balance, contract: %v, address: %v\n", contract, address)
	data := make([]byte, 36)
	copy(data[:4], erc20CodeParts["balanceOf"])
	copy(data[4:], common.HexToAddress(address).Hash().Bytes())
	to := common.HexToAddress(contract)
	msg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	result, err := callContract(client, msg)
	if err != nil {
		return big.NewFloat(0), err
	}
	b := fmt.Sprintf("0x%v", hex.EncodeToString(result))
	ret, _ := com.GetBigIntFromStr(b)
	retf := new(big.Float).SetInt(ret)
	decimal, errd := getTokenDecimal(client, contract)
	if errd != nil {
		return big.NewFloat(0), errd
	}
	retf = new(big.Float).Quo(retf, decimal)
	//fmt.Printf("getErc20Balance, retb: %v\n", retf)
	return retf, nil
}

func callContract(client *ethclient.Client, msg ethereum.CallMsg) ([]byte, error) {
	var i int
	var result []byte
	if client == nil {
		fmt.Printf("callContract, client is nil.\n")
		return []byte{}, com.ErrNoValueObtaind
	}
	var err error
	for i = 0; i < getTokenCount; i++ {
		result, err = client.CallContract(context.Background(), msg, nil)
		if err == nil {
			return result, nil
		}
	}
	return []byte{}, com.ErrNoValueObtaind
}

// getTokenTotalSupply get token total supply
func getErc20TotalSupply(client *ethclient.Client, contract string) (*big.Float, error) {
	//fmt.Printf("getTokenTotalSupply, contract: %v\n", contract)
	data := make([]byte, 4)
	copy(data[:4], erc20CodeParts["totalsupply"])
	to := common.HexToAddress(contract)
	msg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	var ok bool
	var result []byte
	var err error
	for i := 0; i < 3; i++ {
		result, err = callContract(client, msg)
		if err == nil {
			ok = true
			break
		}
		fmt.Printf("getErc20TotalSupply, msg: %v, err: %v\n", msg, err)
	}
	if !ok {
		return big.NewFloat(0), err
	}
	b := fmt.Sprintf("0x%v", hex.EncodeToString(result))
	ret, _ := com.GetBigIntFromStr(b)
	retf := new(big.Float).SetInt(ret)
	decimal, err := getTokenDecimal(client, contract)
	if err != nil {
		return big.NewFloat(0), err
	}
	retf = new(big.Float).Quo(retf, decimal)
	//fmt.Printf("GetTotenTotalSupply, retb: %v\n", retf)
	return retf, nil
}

// getTokenDecimal get token decimal
func getTokenDecimal(client *ethclient.Client, contract string) (*big.Float, error) {
	if Decimal[contract] != nil {
		return Decimal[contract], nil
	}
	//fmt.Printf("getTokenDecimal, contract: %v\n", contract)
	data := make([]byte, 4)
	copy(data[:4], erc20CodeParts["decimal"])
	to := common.HexToAddress(contract)
	msg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	result, err := callContract(client, msg)
	if err != nil {
		return big.NewFloat(0), err
	}
	b := fmt.Sprintf("0x%v", hex.EncodeToString(result))
	decimal, _ := com.GetBigIntFromStr(b)
	Decimal[contract] = big.NewFloat(math.Pow(10, float64(decimal.Int64())))
	//fmt.Printf("getTokenDecimal, retb: %v\n", Decimal)
	return Decimal[contract], nil
}

