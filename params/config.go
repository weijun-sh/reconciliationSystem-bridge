package params

import (
	"encoding/json"
	//"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/weijun-sh/reconciliationSystem-bridge/common"
	bhttp "github.com/weijun-sh/reconciliationSystem-bridge/http"
	"github.com/weijun-sh/reconciliationSystem-bridge/log"
	"github.com/davecgh/go-spew/spew"
)

const TEST bool = false

var (
	recoConfig        *RecoConfig
	priceConfig       *PriceConfig
	loadConfigStarter sync.Once
	loadPriceConfigStarter sync.Once
	bridgeList        BridgeList
)

var (
	Gateway map[string]string = make(map[string]string)
	ChainId map[string]string = make(map[string]string)
	Price map[string]float64 = make(map[string]float64)
	AnyToken map[string]string = make(map[string]string)
	DespositAddress map[string][]string = make(map[string][]string)
	SpecialTotalSupply2Balance map[string]bool = make(map[string]bool)
	PriceExclude map[string]bool = make(map[string]bool)
	CompensationAmount map[string]float64 = make(map[string]float64)
)

func init() {
	ChainId["1"] = "ETH"
	ChainId["5"] = "GOERLI"
	ChainId["56"] = "BSC"
	ChainId["57"] = "SYSCOIN"
	ChainId["66"] = "OKT"
	ChainId["40"] = "TLOS"
	ChainId["100"] = "XDAI"
	ChainId["122"] = "FUSE"
	ChainId["128"] = "HECO"
	ChainId["1285"] = "MOON"
	ChainId["137"] = "MATIC"
	ChainId["250"] = "FTM"
	ChainId["321"] = "KCS"
	ChainId["336"] = "SHI"
	ChainId["4689"] = "IOTEX"
	ChainId["42161"] = "ARB"
	ChainId["42220"] = "CELO"
	ChainId["43114"] = "AVAX"
	ChainId["32659"] = "FSN"
	ChainId["1666600000"] = "HARMONY"
	ChainId["BTC"] = "BTC"
	ChainId["BLOCK"] = "BLOCK"
	ChainId["COLX"] = "COLX"
	ChainId["LTC"] = "LTC"
	ChainId["TERRA"] = "TERRA"

	AnyToken["ETH2FTM-USDCoin"] = "0x7ea2be2df7ba6e54b1a9c70676f668455e329d29"
	AnyToken["ETH2FTM-TetherUSD"] = "0x22648c12acd87912ea1710357b1302c6a4154ebc"
	AnyToken["ETH2FTM-DaiStablecoin"] = "0x739ca6d71365a08f584c8fc4e1029045fa8abc4b"
	AnyToken["ETH2FTM-WrappedEther"] = "0xb153fb3d196a8eb25522705560ac152eeec57901"

	DespositAddress["ETH2FTM"] = []string{"0x5E583B6a1686f7Bc09A6bBa66E852A7C80d36F00"}
	DespositAddress["ETH2BSC"] = []string{"0x39C1Bb10bead88D246aFd96c75a104E09302d7e1"} // v3
	DespositAddress["ETH2BSCv2"] = []string{"0xc5197fd1422fB6c9ea7CdA7Ff5C61Fc352aD6031"}//,"0x533e3c0e6b48010873B947bddC4721b1bDFF9648"} // v2
	DespositAddress["BSC2FTM"] = []string{"0xe1A5B6493054D36DDaC337c2B2f407423Bf08a9F"}
	DespositAddress["BSC2MATIC"] = []string{"0xCf9f53e2222F46024B7C632689804c024C13d03c"}
	DespositAddress["FSN2BSC"] = []string{"0x94e840798e333cB1974E086B58c10C374E966bc7"}

	SpecialTotalSupply2Balance["FTM2ETH-Fantom"] = true
	PriceExclude["FTM2ETH-Fantom"] = true

	CompensationAmount["ETH2MOON-USDCoin"] = 50000.0 // return
}

// LoadConfig load config
func LoadConfig(configFile string) *RecoConfig {
	loadConfigStarter.Do(func() {
		if configFile == "" {
			log.Fatalf("LoadConfig error: no config file specified")
		}
		//log.Println("Config file is", configFile)
		if !common.FileExist(configFile) {
			log.Fatalf("LoadConfig error: config file %v not exist", configFile)
		}
		config := &RecoConfig{}
		if _, err := toml.DecodeFile(configFile, &config); err != nil {
			log.Fatalf("LoadConfig error (toml DecodeFile): %v", err)
		}

		//log.Info("LoadConfig", "config", config)
		SetConfig(config)
		initGateway(config)
		if TEST {
			var bs []byte
			if log.JSONFormat {
				bs, _ = json.Marshal(config)
			} else {
				bs, _ = json.MarshalIndent(config, "", "  ")
			}
			log.Println("LoadConfig finished.", string(bs))
		}
		//TODO test
		bridgeList = getBridgeList(config.BridgeInfo.NetAPI)
	})
	return recoConfig
}

// LoadPriceConfig load config
func LoadPriceConfig(configFile string) *PriceConfig {
	loadPriceConfigStarter.Do(func() {
		if configFile == "" {
			log.Fatalf("LoadPriceConfig error: no price config file specified")
		}
		//log.Println("Config file is", configFile)
		if !common.FileExist(configFile) {
			log.Fatalf("LoadConfig error: config file %v not exist", configFile)
		}
		config := &PriceConfig{}
		if _, err := toml.DecodeFile(configFile, &config); err != nil {
			log.Fatalf("LoadConfig error (toml DecodeFile): %v", err)
		}

		//log.Info("LoadConfig", "config", config)
		SetPriceConfig(config)
		if TEST {
			var bs []byte
			if log.JSONFormat {
				bs, _ = json.Marshal(config)
			} else {
				bs, _ = json.MarshalIndent(config, "", "  ")
			}
			log.Println("LoadPriceConfig finished.", string(bs))
		}
	})
	return priceConfig
}

func initGateway(config *RecoConfig) {
	for _, gateway := range config.Gateway.URL {
		gt := strings.Split(gateway, ",")
		Gateway[gt[0]] = gt[1]
	}
}

func getBridgeList(url string) BridgeList {
	bridgeInfo := bhttp.HttpGet(url)
	info := BridgeInfo{}
	err := json.Unmarshal([]byte(bridgeInfo), &info)
	if err != nil {
		//fmt.Printf("json.Unmarrshal err: %v\n", err)
	}
	bl := BridgeList{}
	bl = info.BridgeList
	sort.Sort(bl)
	if TEST {
		spew.Println(bl)
	}
	return bl
}

func GetBridgeMinTvl() float64 {
	return recoConfig.BridgeInfo.MinTVL
}

// GetConfig get config items structure
func GetConfig() *RecoConfig {
	return recoConfig
}

// SetConfig set config items
func SetConfig(config *RecoConfig) {
	recoConfig = config
}

func GetBridgeList() BridgeList {
	return bridgeList
}

// BridgeConfig config items (decode from toml file)
type RecoConfig struct {
	BridgeInfo *BridgeInfoConfig
	Gateway    *GateWayConfig
	StableCoin *[]StableCoinConfig
}

type BridgeInfoConfig struct {
	NetAPI string
	MinTVL float64
}

type GateWayConfig struct {
	URL []string
}

type StableCoinConfig struct {
	Name   string
	Stolen float64
	Token  []string
}

type BridgeInfo struct {
	BridgeList []*BridgeConfig
}

type BridgeConfig struct {
	ChainId    string
	SrcChainId string
	Token      string
	SrcToken   string
	Symbol     string
	Decimals   uint64
	Name string
	DepositAddr string
	IsProxy bool
	DelegateToken string
	Price float64
	//Sortid uint64
	Type string
	//Balance string
	Tvl float64
}

// SetPriceConfig set config items
func SetPriceConfig(config *PriceConfig) {
	priceConfig = config
	for _, p := range config.SymbolPrice {
		Price[p.ChainSymbol] = p.Price
	}
}

// GetPriceConfig set config items
func GetpriceConfig() *PriceConfig {
	return priceConfig
}

// PriceConfig config items (decode from toml file)
type PriceConfig struct {
	PriceInfo priceInfoConfig
	SymbolPrice []*symbolPriceConfig
}

type priceInfoConfig struct {
	From string
	Time string
}

type symbolPriceConfig struct {
	ChainSymbol string
	Id string
	Price float64
}


