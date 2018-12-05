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
type sdkHandler struct {
	client   *FabricClient
	identity *Identity
}

var (
	logger          = logging.MustGetLogger("sdk")
	handler         sdkHandler
	orgPeerMap      = make(map[string][]string)
	orderNames      []string
	peerNames       []string
	eventName       string
	orRulePeerNames []string
)

func InitSDK(configPath string) error {
	// initialize Fabric client
	var err error
	clientConfig, err := NewClientConfig(configPath)
	if err != nil {
		return err
	}

	if err := SetLogLevel(clientConfig.LogLevel, "sdk"); err != nil {
		return fmt.Errorf("setLogLevel err: %s\n", err.Error())
	}
	logger.Debugf("************InitSDK************by: %s", configPath)

	handler.client, err = NewFabricClientFromConfig(*clientConfig)
	if err != nil {
		return err
	}
	mspPath := handler.client.Channel.MspConfigPath
	if mspPath == "" {
		return fmt.Errorf("config mspPath is empty")
	}
	cert, prikey, err := FindCertAndKeyFile(mspPath)
	if err != nil {
		return err
	}
	handler.identity, err = LoadCertFromFile(cert, prikey)
	if err != nil {
		return err
	}
	handler.identity.MspId = handler.client.Channel.LocalMspId

	if err := parsePolicy(); err != nil {
		return fmt.Errorf("parsePolicy err: %s\n", err.Error())
	}
	return err
}

// GetHandler get sdk handler
func GetHandler() *sdkHandler {
	return &handler
}

// GetHandler get sdk handler
func GetConfigLogLevel() string {
	return handler.client.Log.LogLevel
}

// Invoke invoke cc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) Invoke(args []string, channelName, chaincodeName string) (*InvokeResponse, error) {
	peerNames := getSendPeerName()
	orderName := getSendOrderName()
	if len(peerNames) == 0 || orderName == "" {
		return nil, fmt.Errorf("config peer order is err")
	}
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName)
	if err != nil {
		return nil, err
	}
	return sdk.client.Invoke(*sdk.identity, *chaincode, peerNames, orderName)
}

// Query query cc  ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) Query(args []string, channelName, chaincodeName string) ([]*QueryResponse, error) {
	peerNames := getSendPeerName()
	if len(peerNames) == 0 {
		return nil, fmt.Errorf("config peer order is err")
	}
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName)
	if err != nil {
		return nil, err
	}

	return sdk.client.Query(*sdk.identity, *chaincode, []string{peerNames[0]})
}

// Query query qscc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) QueryByQscc(args []string, channelName string) ([]*QueryResponse, error) {
	peerNames := getSendPeerName()
	if len(peerNames) == 0 {
		return nil, fmt.Errorf("config peer order is err")
	}

	mspId := handler.client.Channel.LocalMspId
	if channelName == "" || mspId == "" {
		return nil, fmt.Errorf("channelName or mspid is empty")
	}

	chaincode := ChainCode{
		ChannelId: channelName,
		Type:      ChaincodeSpec_GOLANG,
		Name:      QSCC,
		Args:      args,
	}

	return sdk.client.Query(*sdk.identity, chaincode, []string{peerNames[0]})
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) GetBlockByNumber(blockNum uint64, channelName string) (*common.Block, error) {
	strBlockNum := strconv.FormatUint(blockNum, 10)
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("GetBlockHeight channelName is empty ")
	}
	args := []string{"GetBlockByNumber", channelName, strBlockNum}
	logger.Debugf("GetBlockByNumber chainId %s num %s", channelName, strBlockNum)
	resps, err := sdk.QueryByQscc(args, channelName)
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
func (sdk *sdkHandler) GetBlockHeight(channelName string) (uint64, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if channelName == "" {
		return 0, fmt.Errorf("GetBlockHeight channelName is empty ")
	}
	args := []string{"GetChainInfo", channelName}
	resps, err := sdk.QueryByQscc(args, channelName)
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
func (sdk *sdkHandler) GetBlockHeightByEventName(channelName string) (uint64, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	args := []string{"GetChainInfo", channelName}
	mspId := handler.client.Channel.LocalMspId
	if channelName == "" || mspId == "" {
		return 0, fmt.Errorf("channelName or mspid is empty")
	}
	if eventName == "" {
		return 0, fmt.Errorf("event peername is empty")
	}
	chaincode := ChainCode{
		ChannelId: channelName,
		Type:      ChaincodeSpec_GOLANG,
		Name:      QSCC,
		Args:      args,
	}

	resps, err := sdk.client.QueryByEvent(*sdk.identity, chaincode, []string{eventName})
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
func (sdk *sdkHandler) ListenEventFullBlock(channelName string, startNum int) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("ListenEventFullBlock channelName is empty ")
	}
	ch := make(chan parseBlock.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFullBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
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
func (sdk *sdkHandler) ListenEventFilterBlock(channelName string, startNum int) (chan EventBlockResponse, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("ListenEventFilterBlock  channelName is empty ")
	}
	ch := make(chan EventBlockResponse)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFilteredBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
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
func (sdk *sdkHandler) Listen(peerName, channelName string) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if channelName == "" {
		return nil, fmt.Errorf("Listen  channelName is empty ")
	}
	mspId := sdk.client.Channel.LocalMspId
	if mspId == "" {
		return nil, fmt.Errorf("Listen  mspId is empty ")
	}
	ch := make(chan parseBlock.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.Listen(ctx, sdk.identity, peerName, channelName, mspId, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	return ch, nil
}

func (sdk *sdkHandler) GetOrdererConnect() (bool, error) {
	orderName := getSendOrderName()
	if orderName == "" {
		return false, fmt.Errorf("config order is err")
	}
	if _, ok := sdk.client.Orderers[orderName]; ok {
		ord := sdk.client.Orderers[orderName]
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

//解析区块
func (sdk *sdkHandler) ParseCommonBlock(block *common.Block) (*parseBlock.Block, error) {
	blockObj := parseBlock.ParseBlock(block, 0)
	return &blockObj, nil
}

// param channel only used for create channel, if upate config channel should be nil
func (sdk *sdkHandler) ConfigUpdate(payload []byte, channel string) error {
	orderName := getSendOrderName()
	if channel != "" {
		return sdk.client.ConfigUpdate(*sdk.identity, payload, channel, orderName)
	}
	return sdk.client.ConfigUpdate(*sdk.identity, payload, sdk.client.Channel.ChannelId, orderName)
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
