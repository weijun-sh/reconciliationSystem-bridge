package tokens

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/weijun-sh/reconciliationSystem-bridge/common"
	"github.com/weijun-sh/reconciliationSystem-bridge/params"
	"github.com/weijun-sh/reconciliationSystem-bridge/rpc"
)

var profitPrice *big.Float

func GetBalanceOfToken() {
	profitPrice = big.NewFloat(0)
	printfHeader()
	minTvl := params.GetBridgeMinTvl()
	bridgeList := params.GetBridgeList()
	j := 0
	for _, bl := range bridgeList {
		if bl.Type != "bridge" { // not bridge
			continue
		}
		if bl.Tvl < minTvl { // Rank greater than min
			break
		}
		j += 1
		balanceTmp, balancePrint := getBridgeBalance(bl)
		totalSupplyTmp, totalSuplyPrint := getBridgeTotalSupply(bl)

		printfBody(bl, j, balanceTmp, balancePrint, totalSupplyTmp, totalSuplyPrint)
	}
	printfTail(j)
}

func getBridgeBalance(bl *params.BridgeConfig) (*big.Float, string) {
	//fmt.Printf("getBridgeBalance, bl: %v\n", bl)
	if bl.SrcToken == "" || bl.SrcToken == "0x0000000000000000000000000000000000000000" || len(bl.SrcToken) < 20 { // native
		srcChain := params.ChainId[bl.SrcChainId]
		if isBTCChain(srcChain) {
			return GetBtcBalance(bl)
		}
		return GetEthBalance(bl)
	} else {
		return GetTokenBalances(bl)
	}
}

func getBridgeTotalSupply(bl *params.BridgeConfig) (*big.Float, string) {
	srcChain := params.ChainId[bl.SrcChainId]
	destChain := params.ChainId[bl.ChainId]
	bridge := fmt.Sprintf("%v2%v-%v", srcChain, destChain, bl.Name)
	if params.SpecialTotalSupply2Balance[bridge] { // FTM2ETH-Fantom
		return GetTokenBalances(bl)
	} else {
		return GetTokenTotalSupply(bl)
	}
}

func GetTokenBalances(bl *params.BridgeConfig) (*big.Float, string) {
	//fmt.Printf("GetTokenBalances, bl: %v\n", bl)
	srcChain := params.ChainId[bl.SrcChainId]
	if srcChain == "" {
		fmt.Printf("GetTokenBalances, srcChainId: %v not set\n", bl.SrcChainId)
	}
	chain := srcChain
	destChain := params.ChainId[bl.ChainId]
	if destChain == "" {
		fmt.Printf("GetTokenBalances, destChainId: %v not set\n", bl.ChainId)
	}
	srcToken := bl.SrcToken
	token := srcToken
	depositAddr := bl.DepositAddr
	bridge := fmt.Sprintf("%v2%v-%v", srcChain, destChain, bl.Name)
	bridgeChain := fmt.Sprintf("%v2%v", srcChain, destChain)
	if bridgeChain == "ETH2BSC" && (bl.Symbol == "SUPER" || bl.Symbol == "MTLX") { // V2
		bridgeChain = fmt.Sprintf("%vv2", bridgeChain)
	}
	if params.SpecialTotalSupply2Balance[bridge] { // FTM2ETH-Fantom
		depositAddr = bl.DelegateToken
		chain = destChain
		token = bl.Token
	}
	balanceTmp, balancePrint := GetTokenBalance(chain, token, depositAddr)
	if len(params.DespositAddress[bridgeChain]) > 0 { // address2
		for _, addr := range params.DespositAddress[bridgeChain] {
			balanceTmp2, _ := GetTokenBalance(chain, token, addr)
			balanceTmp = new(big.Float).Add(balanceTmp, balanceTmp2)
			balancePrint = fmt.Sprintf("%0.2f", balanceTmp)
		}
	}
	if len(params.AnyToken[bridge]) == 42 { // anyToken
		anyToken := params.AnyToken[bridge]
		balanceTmp2, _ := GetTokenBalance(chain, anyToken, depositAddr)
		balanceTmp = new(big.Float).Add(balanceTmp, balanceTmp2)
		balancePrint = fmt.Sprintf("%0.2f", balanceTmp)
		if len(params.DespositAddress[bridgeChain]) > 0 { // address2
			for _, addr := range params.DespositAddress[bridgeChain] {
				balanceTmp2, _ := GetTokenBalance(chain, anyToken, addr)
				balanceTmp = new(big.Float).Add(balanceTmp, balanceTmp2)
				balancePrint = fmt.Sprintf("%0.2f", balanceTmp)
			}
		}
	}
	//CompensationAmount
	if params.CompensationAmount[bridge] > 0.0 {
		c := new(big.Float).SetFloat64(params.CompensationAmount[bridge])
		balanceTmp = new(big.Float).Add(balanceTmp, c)
	}
	return balanceTmp, balancePrint
}

func isBTCChain(chain string) bool {
	if chain == "BTC" || chain == "LTC" || chain == "COLX" {
		return true
	}
	return false
}

func GetTokenBalance(chain, contract, addr string) (*big.Float, string) {
	balance, err := getTokenBalance(chain, contract, addr)
	balanceP := ""
	switch {
	case errors.Is(err, common.ErrAddressNull):
		balanceP = "*addrNull"
	case errors.Is(err, common.ErrAddressInValid):
		balanceP = "*addrInValid"
	case errors.Is(err, nil):
		balanceP = fmt.Sprintf("%0.2f", balance)
	}
	//fmt.Printf("GetTokenBalance, contract: %v, addr: %v, balance: %0.2f\n", contract, addr, balance)
	return balance, balanceP
}

func checkAddressValid(addr string) error {
	if len(addr) == 0 {
		return common.ErrAddressNull
	}
	if !common.IsHexAddress(addr) {
		return common.ErrAddressInValid
	}
	return nil
}
func getTokenBalance(chain, contract, addr string) (*big.Float, error) {
	err := checkAddressValid(contract)
	if err != nil {
		return big.NewFloat(0), err
	}
	err = checkAddressValid(addr)
	if err != nil {
		return big.NewFloat(0), err
	}
	client := rpc.ClientRpc.Client[chain]
	return getErc20Balance(client, contract, addr)
}

func GetTokenTotalSupply(bl *params.BridgeConfig) (*big.Float, string) {
	destChain := params.ChainId[bl.ChainId]
	if destChain == "" {
		fmt.Printf("GetTokenTotalSupply(), destChain(%v) not config\n", bl.ChainId)
		return big.NewFloat(0), ""
	}
	destToken := bl.Token
	totalSupplyTmp, err := getTokenTotalSupply(destChain, destToken)
	totalSuplyPrint := ""
	switch {
	case errors.Is(err, common.ErrAddressNull):
		totalSuplyPrint = "*addrNull"
	case errors.Is(err, common.ErrAddressInValid):
		totalSuplyPrint = "*addrInvalid"
	case errors.Is(err, nil):
		totalSuplyPrint = fmt.Sprintf("%0.2f", totalSupplyTmp)
	}
	return totalSupplyTmp, totalSuplyPrint
}

func getTokenTotalSupply(chain, addr string) (*big.Float, error) {
	//fmt.Printf("getTokenTotalSupply, chain: %v, addr: %v\n", chain, addr)
	err := checkAddressValid(addr)
	if err != nil {
		return big.NewFloat(0), err
	}
	client := rpc.ClientRpc.Client[chain]
	if client == nil {
		fmt.Printf("getTokenTotalSupply, chain: %v, rpc client is nil\n", chain)
		return big.NewFloat(0), errors.New("rpc client is nil")
	}
	return getErc20TotalSupply(client, addr)
}

func GetEthBalance(bl *params.BridgeConfig) (*big.Float, string) {
	srcChain := params.ChainId[bl.SrcChainId]
	destChain := params.ChainId[bl.ChainId]
	depositAddr := bl.DepositAddr
	bridgeChain := fmt.Sprintf("%v2%v", srcChain, destChain)
	if bridgeChain == "ETH2BSC" && (bl.Symbol == "SUPER" || bl.Symbol == "MTLX") {
		bridgeChain = fmt.Sprintf("%vv2", bridgeChain)
	}
	url := params.Gateway[srcChain]
	var balanceTmp *big.Float
	var balancePrint string
	if isChainTerra(srcChain) {
		balanceTmp, balancePrint = getBalance4TERRA(url, depositAddr, bl.Symbol)
	} else {
		balanceTmp, balancePrint = getBalance4ETH(url, depositAddr)
	}
	//fmt.Printf("GetETHBalance, bridgeChain: %v, addr: %v, balance: %0.2f\n", bridgeChain, depositAddr, balanceTmp)
	if len(params.DespositAddress[bridgeChain]) > 0 { // address2
		for _, addr := range params.DespositAddress[bridgeChain] {
			var balanceTmp2 *big.Float
			if isChainTerra(srcChain) {
				balanceTmp2, _ = getBalance4TERRA(url, addr, bl.Symbol)
			} else {
				balanceTmp2, _ = getBalance4ETH(url, addr)
			}
			//fmt.Printf("GetETHBalance, bridgeChain: %v, addr: %v, balance: %0.2f\n", bridgeChain, addr, balanceTmp2)
			balanceTmp = new(big.Float).Add(balanceTmp, balanceTmp2)
			balancePrint = fmt.Sprintf("%0.2f", balanceTmp)
		}
	}
	return balanceTmp, balancePrint
}

func isChainTerra(chain string) bool {
	return strings.EqualFold(chain, "TERRA")
}

func GetBtcBalance(bl *params.BridgeConfig) (*big.Float, string) {
	srcChain := params.ChainId[bl.SrcChainId]
	destChain := params.ChainId[bl.ChainId]
	depositAddr := bl.DepositAddr
	bridgeChain := fmt.Sprintf("%v2%v", srcChain, destChain)
	url := params.Gateway[srcChain]
	balanceTmp, balancePrint := getBalance4BTC(url, depositAddr)
	if len(params.DespositAddress[bridgeChain]) > 0 { // address2
		for _, addr := range params.DespositAddress[bridgeChain] {
			balanceTmp2, _ := getBalance4BTC(url, addr)
			balanceTmp = new(big.Float).Add(balanceTmp, balanceTmp2)
			balancePrint = fmt.Sprintf("%0.2f", balanceTmp)
		}
	}
	return balanceTmp, balancePrint
}

// print
func printfHeader() {
	minTvl := params.GetBridgeMinTvl()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("ps. update every 20 minutes\n")
	fmt.Printf("ps. TVL(USD) >= %0.2f, sort descending\n", minTvl)
	fmt.Printf("===================================================================================================================================\n")
	fmt.Printf("                                                 BRIDGE COIN RECONCILIATION SYSTEM                                                 \n")
	fmt.Printf("===================================================================================================================================\n")
	fmt.Printf("    | name                  | srcChain chain UnderlyingBalance  TotalSupply |          profit |           price |      total(USD)\n")
	fmt.Printf("----+-----------------------+-----------------------------------------------+ ----------------+-----------------+------------------\n")
}

func printfBody(bl *params.BridgeConfig, i int, balanceTmp *big.Float, balancePrint string, totalSupplyTmp *big.Float, totalSuplyPrint string) {
	srcChain := params.ChainId[bl.SrcChainId]
	destChain := params.ChainId[bl.ChainId]
	bridge := fmt.Sprintf("%v2%v-%v", srcChain, destChain, bl.Name)
	profit := big.NewFloat(0)
	isLessThanZero := false
	if params.SpecialTotalSupply2Balance[bridge] { // FTM2ETH-Fantom
		profit = new(big.Float).Add(balanceTmp, totalSupplyTmp)
	} else {
		profit = new(big.Float).Sub(balanceTmp, totalSupplyTmp)
		if balanceTmp.Cmp(totalSupplyTmp) < 0{
			isLessThanZero = true
		}
	}

	// price
	profitPricePrintf := "-"
	symbol := strings.ToLower(bl.Symbol)
	sp := fmt.Sprintf("%v-%v-%v", srcChain, destChain, symbol)
	price := params.Price[sp]
	if price <= 0.0 {
		fmt.Printf("price is 0.0, sp: %v, srcToken: %v, Token: %v, bl.Symbol: %v\n", sp, bl.SrcToken, bl.Token, bl.Symbol)
	}
	pricePrintf := fmt.Sprintf("%0.4f", price)
	if price < 0.0001 {
		pricePrintf = fmt.Sprintf("%v", price)
	}
	profitPriceTmp := big.NewFloat(0)
	if !params.PriceExclude[bridge] && profit.Cmp(big.NewFloat(0)) > 0 && price > 0.0 {
		priceBig := new(big.Float).SetFloat64(price)
		profitPriceTmp = new(big.Float).Mul(profit, priceBig)
		profitPricePrintf = fmt.Sprintf("%0.2f", profitPriceTmp)
	}

	profitPrintf := fmt.Sprintf("%0.2f", profit)
	//fmt.Printf("srcChainId: %v, chainId: %v, token: %v, symbol: %v, price: %v, name: %v, srcToken: %v\n", bl.SrcChainId, bl.ChainId, bl.Token, bl.Symbol, bl.Price, bl.Name, bl.SrcToken)
	fmt.Printf("%3v | %-21v | ", i, bl.Name)
	fmt.Printf("%5v %5v %16v %16v | %15v | %15v | %15v", srcChain, destChain, balancePrint, totalSuplyPrint, profitPrintf, pricePrintf, profitPricePrintf)
	if isLessThanZero {
		fmt.Printf("      *")
	}
	profitPrice = new(big.Float).Add(profitPrice, profitPriceTmp)
	fmt.Println()
}

func printfTail(i int) {
	fmt.Printf("===================================================================================================================================\n")
	profitPricePrintf := fmt.Sprintf("%0.2f", profitPrice)
	fmt.Printf("%3v                                                                                                       total   %15v (USD)\n", i, profitPricePrintf)
}

