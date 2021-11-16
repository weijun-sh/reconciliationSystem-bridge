package rpc

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/weijun-sh/reconciliationSystem-bridge/params"
)

var ClientRpc *ClientRPC = &ClientRPC{}

func InitClient() *ClientRPC {
	ClientRpc.Client = make(map[string]*ethclient.Client, 0)

	config := params.GetConfig()
	for _, gateway := range config.Gateway.URL {
		gt := strings.Split(gateway, ",")
		ClientRpc.Client[gt[0]] = initClient(gt[1])
	}
	return ClientRpc
}

//			getErc20Balance(ethcli, tokenPair.UnderlyingAddr, tokenPair.TokenAddr)
func initClient(gateway string) *ethclient.Client {
	ethcli, err := ethclient.Dial(gateway)
	if err != nil {
		fmt.Printf("ethclient.Dail failed, gateway: %v, err: %v\n", gateway, err)
		return nil
	}
	//fmt.Printf("ethclient.Dail gateway success, gateway: %v\n", gateway)
	return ethcli
}

type ClientRPC struct {
	Client map[string]*ethclient.Client
}
