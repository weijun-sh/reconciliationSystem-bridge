package tokens

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	//"net/url"
	//"unsafe"
	"io/ioutil"
	"math"
	"math/big"
)

type ethConfig struct {
	Result string `bson: "result"`
}

func getBalance4ETH(url, address string) (*big.Float, string) {
	//fmt.Printf("getBalance4ETH, url: %v, address: %v\n", url, address)
	data := make(map[string]interface{})
	data["method"] = "eth_getBalance"
	data["params"] = []string{address, "latest"}
	data["id"] = "1"
	data["jsonrpc"] = "2.0"
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return big.NewFloat(0), "~"
	}
	reader := bytes.NewReader(bytesData)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		fmt.Println(err.Error())
		return big.NewFloat(0), "~"
	}
	defer resp.Body.Close()

	//fmt.Printf("r: %v\n", resp)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return big.NewFloat(0), "~"
	}
	basket := ethConfig{}
	err = json.Unmarshal(body, &basket)
	if err != nil {
		fmt.Println(err)
		return big.NewFloat(0), "~"
	}
	//fmt.Printf("b: %v\n", basket)
	b := getBalance4String(basket.Result, 18)
	bs := fmt.Sprintf("%0.2f", b)
	return b, bs
}

func getBalance4String(balance string, d int) *big.Float {
	fbalance := new(big.Float)
	fbalance.SetString(balance)
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(d)))
	return ethValue
}
