package worker

import (
	//"time"
	"github.com/weijun-sh/reconciliationSystem-bridge/rpc"
	"github.com/weijun-sh/reconciliationSystem-bridge/tokens"
)

// StartWork start get balance
func StartWork() {
	rpc.InitClient()

	tokens.GetBalanceOfToken()
}
