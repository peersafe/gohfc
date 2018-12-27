package gohfc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc/parseBlock"
	"google.golang.org/grpc/connectivity"
	"strconv"
)

//sdk handler
type SdkHandler struct {
	FabCli  *FabricClient
	IdeCli  *Identity
	Yconfig *YamlConfig
}

var (
	logger = logging.MustGetLogger("sdk")
)

func InitSdkByFile(configPath string) (*SdkHandler, error) {
	// initialize Fabric FabCli
	var err error
	yamlConfig, err := NewYamlConfig(configPath)
	if err != nil {
		return nil, err
	}

	logger.Infof("************InitSdkByFile************by: %s", configPath)
	fabclient, err := NewFabricClientFromConfig(yamlConfig)
	if err != nil {
		return nil, err
	}

	cert, prikey, err := FindCertAndKeyFile(yamlConfig.MspConfigPath)
	if err != nil {
		return nil, err
	}
	ideclient, err := LoadCertFromFile(cert, prikey)
	if err != nil {
		return nil, err
	}
	ideclient.MspId = yamlConfig.LocalMspId

	return &SdkHandler{fabclient, ideclient, yamlConfig}, err
}

// Invoke invoke cc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) Invoke(args []string, channelName, chaincodeName string) (*InvokeResponse, error) {
	chaincode, err := sdk.getChainCodeObj(args, channelName, chaincodeName)
	if err != nil {
		return nil, err
	}
	return sdk.FabCli.Invoke(*sdk.IdeCli, *chaincode, sdk.FabCli.getAllPeerNames(), sdk.FabCli.getOneOrdererName())
}

// Query query cc  ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) Query(args []string, channelName, chaincodeName string) ([]*QueryResponse, error) {
	chaincode, err := sdk.getChainCodeObj(args, channelName, chaincodeName)
	if err != nil {
		return nil, err
	}

	return sdk.FabCli.Query(*sdk.IdeCli, *chaincode, []string{sdk.FabCli.getOnePeerName()})
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) GetBlockByNumber(blockNum uint64, channelName string) (*common.Block, error) {
	strBlockNum := strconv.FormatUint(blockNum, 10)
	args := []string{"GetBlockByNumber", channelName, strBlockNum}
	logger.Debugf("GetBlockByNumber chainId %s num %s", channelName, strBlockNum)
	resps, err := sdk.Query(args, channelName, QSCC)
	if err != nil {
		return nil, fmt.Errorf("can not get installed chaincodes :%s", err.Error())
	} else if len(resps) == 0 {
		return nil, fmt.Errorf("GetBlockByNumber empty response from peer")
	}
	if resps[0].Error != nil {
		return nil, resps[0].Error
	}
	data := resps[0].Response.Response.Payload
	var block = new(common.Block)
	err = proto.Unmarshal(data, block)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber Unmarshal from payload failed: %s", err.Error())
	}

	return block, nil
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) GetBlockHeight(channelName string) (uint64, error) {
	args := []string{"GetChainInfo", channelName}
	resps, err := sdk.Query(args, channelName, QSCC)
	if err != nil {
		return 0, err
	} else if len(resps) == 0 {
		return 0, fmt.Errorf("GetChainInfo is empty respons from peer qscc")
	}

	if resps[0].Error != nil {
		return 0, resps[0].Error
	}

	data := resps[0].Response.Response.Payload
	var chainInfo = new(common.BlockchainInfo)
	err = proto.Unmarshal(data, chainInfo)
	if err != nil {
		return 0, fmt.Errorf("GetChainInfo unmarshal from payload failed: %s", err.Error())
	}
	return chainInfo.Height, nil
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) GetBlockHeightByEventName(channelName string) (uint64, error) {
	if len(channelName) == 0 {
		channelName = sdk.Yconfig.ChannelId
	}
	if channelName == "" {
		return 0, fmt.Errorf("GetBlockHeightByEventName channelName is empty")
	}
	args := []string{"GetChainInfo", channelName}
	chaincode := ChainCode{
		ChannelId: channelName,
		Type:      ChaincodeSpec_GOLANG,
		Name:      QSCC,
		Args:      args,
	}

	resps, err := sdk.FabCli.Query(*sdk.IdeCli, chaincode, []string{sdk.FabCli.getOnePeerName()})
	if err != nil {
		return 0, err
	} else if len(resps) == 0 {
		return 0, fmt.Errorf("GetBlockHeightByEventName is empty respons from peer qscc")
	}

	if resps[0].Error != nil {
		return 0, resps[0].Error
	}

	data := resps[0].Response.Response.Payload
	var chainInfo = new(common.BlockchainInfo)
	err = proto.Unmarshal(data, chainInfo)
	if err != nil {
		return 0, fmt.Errorf("GetBlockHeightByEventName unmarshal from payload failed: %s", err.Error())
	}
	return chainInfo.Height, nil
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) ListenEventFullBlock(channelName string, startNum int) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.Yconfig.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("ListenEventFullBlock channelName is empty ")
	}
	ch := make(chan parseBlock.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.FabCli.ListenForFullBlock(ctx, *sdk.IdeCli, startNum, sdk.FabCli.getOnePeerName(), channelName, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	//
	//for d := range ch {
	//	fmt.Println(d)
	//}
	return ch, nil
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *SdkHandler) ListenEventFilterBlock(channelName string, startNum int) (chan EventBlockResponse, error) {
	if len(channelName) == 0 {
		channelName = sdk.Yconfig.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("ListenEventFilterBlock  channelName is empty ")
	}

	ch := make(chan EventBlockResponse)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.FabCli.ListenForFilteredBlock(ctx, *sdk.IdeCli, startNum, sdk.FabCli.getOnePeerName(), channelName, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	//
	//for d := range ch {
	//	fmt.Println(d)
	//}
	return ch, nil
}

//if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
// Listen v 1.0.4 -- port ==> 7053
func (sdk *SdkHandler) Listen(peerName, channelName string) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.Yconfig.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("Listen  channelName is empty ")
	}
	mspId := sdk.Yconfig.LocalMspId
	if mspId == "" {
		return nil, fmt.Errorf("Listen  mspId is empty ")
	}
	ch := make(chan parseBlock.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.FabCli.Listen(ctx, sdk.IdeCli, peerName, channelName, mspId, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	return ch, nil
}

func (sdk *SdkHandler) GetOrdererConnect() (bool, error) {
	orderName := sdk.FabCli.getOneOrdererName()
	if orderName == "" {
		return false, fmt.Errorf("config order is err")
	}
	if _, ok := sdk.FabCli.Orderers[orderName]; ok {
		ord := sdk.FabCli.Orderers[orderName]
		if ord != nil && ord.con != nil {
			if ord.con.GetState() == connectivity.Ready {
				return true, nil
			} else {
				return false, fmt.Errorf("the orderer connect state %s:%s", orderName, ord.con.GetState().String())
			}
		} else {
			return false, fmt.Errorf("the orderer or connect is nil")
		}
	} else {
		return false, fmt.Errorf("the orderer %s is not match", orderName)
	}
}

func (sdk *SdkHandler) getChainCodeObj(args []string, channelName, chaincodeName string) (*ChainCode, error) {
	if len(channelName) == 0 {
		channelName = sdk.Yconfig.ChannelId
	}
	if len(chaincodeName) == 0 {
		chaincodeName = sdk.Yconfig.ChaincodeName
	}
	if channelName == "" {
		return nil, fmt.Errorf("channelName is empty")
	}
	if chaincodeName == "" {
		return nil, fmt.Errorf(" chaincodeName is empty")
	}

	chaincode := ChainCode{
		ChannelId: channelName,
		Type:      ChaincodeSpec_GOLANG,
		Name:      chaincodeName,
		Args:      args,
	}

	return &chaincode, nil
}

//解析区块
func (sdk *SdkHandler) ParseCommonBlock(block *common.Block) (*parseBlock.Block, error) {
	blockObj := parseBlock.ParseBlock(block, 0)
	return &blockObj, nil
}

// param channel only used for create channel, if upate config channel should be nil
func (sdk *SdkHandler) ConfigUpdate(payload []byte, channel string) error {
	orderName := sdk.FabCli.getOneOrdererName()
	if channel != "" {
		return sdk.FabCli.ConfigUpdate(*sdk.IdeCli, payload, channel, orderName)
	}
	return sdk.FabCli.ConfigUpdate(*sdk.IdeCli, payload, sdk.Yconfig.ChannelId, orderName)
}

type KeyValue struct {
	Key   string `json:"key"`   //存储数据的key
	Value string `json:"value"` //存储数据的value
}

func SetArgsTxid(txid string, args *[]string) {
	if len(*args) == 2 && (*args)[0] == "SaveData" {
		var invokeRequest KeyValue
		if err := json.Unmarshal([]byte((*args)[1]), &invokeRequest); err != nil {
			logger.Debugf("SetArgsTxid umarshal invokeRequest failed")
			return
		}
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(invokeRequest.Value), &msg); err != nil {
			logger.Debugf("SetArgsTxid umarshal message failed")
			return
		}
		invokeRequest.Key = txid
		msg["fabricTxId"] = txid
		v, _ := json.Marshal(msg)
		invokeRequest.Value = string(v)
		tempData, _ := json.Marshal(invokeRequest)
		//logger.Debugf("SetArgsTxid msg is %s", tempData)
		(*args)[1] = string(tempData)
	}
}
