package params

//"sort"

type BridgeList []*BridgeConfig

func (b BridgeList) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BridgeList) Len() int           { return len(b) }
func (b BridgeList) Less(i, j int) bool { return b[i].Tvl > b[j].Tvl }
