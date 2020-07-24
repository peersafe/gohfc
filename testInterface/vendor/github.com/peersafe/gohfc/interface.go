package gohfc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc/parseBlock"
	"github.com/peersafe/gohfc/waitTxstatus"
	"google.golang.org/grpc/connectivity"
	"reflect"
	"strconv"
	"strings"
	"time"
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

type FuncGetRAMBLock func() (uint64, error)

// LocalProvider returns local client context
type FuncListenFilterBlock func(channelName string, startNum int, response chan<- EventBlockResponse) error
type FuncListenFullBlock func(channelName string, startNum int, response chan<- parseBlock.Block) error

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
	if handler.client.Channel.ChannelId == "" {
		return fmt.Errorf("config channelid is empty")
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
	//check sdk crypto type don`t match private key
	if key, ok := handler.identity.PrivateKey.(*ecdsa.PrivateKey); ok {
		switch key.Curve {
		case elliptic.P256(), elliptic.P384(), elliptic.P521():
			if strings.ToUpper(clientConfig.CryptoConfig.Family) != "ECDSA" {
				return fmt.Errorf("the sdk crypto type %s don`t match private key", clientConfig.CryptoConfig.Family)
			}
		default:
			if strings.ToUpper(clientConfig.CryptoConfig.Family) == "ECDSA" {
				return fmt.Errorf("the sdk crypto type %s don`t match private key", clientConfig.CryptoConfig.Family)
			}
		}
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

// GetHandler get sdk handler
func GetChaincodeName() string {
	return handler.client.Channel.ChaincodeName
}

// Sync Invoke invoke cc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) SyncInvoke(args []string, channelName, chaincodeName string) (*InvokeResponse, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	} else if channelName != sdk.client.Channel.ChannelId {
		return nil, fmt.Errorf("%s, %s dont`t match, no support sync invoke", channelName, sdk.client.Channel.ChannelId)
	}
	peerNames := getSendPeerName()
	orderName := getSendOrderName()
	if len(peerNames) == 0 || orderName == "" {
		return nil, fmt.Errorf("config peer order is err")
	}
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName, "")
	if err != nil {
		return nil, err
	}
	return sdk.client.SyncInvoke(*sdk.identity, *chaincode, peerNames, orderName)
}

// Invoke invoke cc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) Invoke(args []string, channelName, chaincodeName string) (*InvokeResponse, error) {
	peerNames := getSendPeerName()
	orderName := getSendOrderName()
	if len(peerNames) == 0 || orderName == "" {
		return nil, fmt.Errorf("config peer order is err")
	}
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName, "")
	if err != nil {
		return nil, err
	}
	return sdk.client.Invoke(*sdk.identity, *chaincode, peerNames, orderName)
}

// Invoke invoke with private data cc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) InvokeWithPriData(args []string, channelName, chaincodeName, pridata string) (*InvokeResponse, error) {
	peerNames := getSendPeerName()
	orderName := getSendOrderName()
	if len(peerNames) == 0 || orderName == "" {
		return nil, fmt.Errorf("config peer order is err")
	}
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName, pridata)
	if err != nil {
		return nil, err
	}
	return sdk.client.Invoke(*sdk.identity, *chaincode, peerNames, orderName)
}

// Query query cc  ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) Query(args []string, channelName, chaincodeName string) ([]*QueryResponse, error) {
	chaincode, err := getChainCodeObj(args, channelName, chaincodeName, "")
	if err != nil {
		return nil, err
	}

	return sdk.client.Query(*sdk.identity, *chaincode, []string{eventName})
}

// Query query qscc ,if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) QueryByQscc(args []string, channelName string) ([]*QueryResponse, error) {
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

	return sdk.client.Query(*sdk.identity, chaincode, []string{eventName})
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) GetBlockByNumber(blockNum uint64, channelName string) (*common.Block, error) {
	strBlockNum := strconv.FormatUint(blockNum, 10)
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
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
func (sdk *sdkHandler) GetTransactionById(txId string, channelName string) (*parseBlock.Transaction, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	if txId == "" {
		return nil, fmt.Errorf("txid is empty")
	}
	args := []string{"GetTransactionByID", channelName, txId}
	logger.Debugf("GetTransactionByID chainId %s txId %s", channelName, txId)
	resps, err := sdk.QueryByQscc(args, channelName)
	if err != nil {
		return nil, fmt.Errorf("can not get installed chaincodes :%s", err.Error())
	} else if len(resps) == 0 {
		return nil, fmt.Errorf("GetTransactionByID empty response from peer")
	}
	if resps[0].Error != nil {
		return nil, resps[0].Error
	}
	data := resps[0].Response.Response.Payload
	tx, _, err := parseBlock.ParseProcessedTransaction(data)
	return tx, err
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) GetBlockHeight(channelName string) (uint64, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
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
func (sdk *sdkHandler) ListenEventFullBlock(channelName string, startNum int) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}

	ch := make(chan parseBlock.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFullBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
	if err != nil {
		cancel()
		return nil, err
	}

	return ch, nil
}

func (sdk *sdkHandler) HandleTxStatus(channelName string) error {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}

	waitTxstatus.HandleRegisterTxStatusEvent()

	filterBlockChan := make(chan EventBlockResponse)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFilteredBlock(ctx, *sdk.identity, -1, eventName, channelName, filterBlockChan)
	if err != nil {
		cancel()
		return err
	}
	RamBlock := func() (uint64, error) {
		return waitTxstatus.GlobalBlockNumber.Get(), nil
	}
	go sdk.JudgeFilterEventConnect(RamBlock, sdk.EventFilterBlock, filterBlockChan)
	go func() {
		for {
			select {
			case filterBlock := <-filterBlockChan:
				waitTxstatus.GlobalBlockNumber.Put(filterBlock.BlockHeight)
				for _, tx := range filterBlock.Transactions {
					waitTxstatus.PublishTxStatus(tx.Id, tx.Status)
				}
			}
		}
	}()
	return nil
}

// if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
func (sdk *sdkHandler) ListenEventFilterBlock(channelName string, startNum int) (chan EventBlockResponse, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}

	ch := make(chan EventBlockResponse)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFilteredBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	return ch, nil
}

//if channelName ,chaincodeName is nil that use by client_sdk.yaml set value
// Listen v 1.0.4 -- port ==> 7053
func (sdk *sdkHandler) Listen(peerName, channelName string) (chan parseBlock.Block, error) {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}

	if peerName == "" {
		peerName = eventName
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
	if reflect.ValueOf(block).IsNil() || block == nil || block.XXX_Size() == 0 {
		return nil, fmt.Errorf("the block not exist")
	}
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
	if len(*args) == 2 && (*args)[0] == "SaveData" && strings.Contains((*args)[1], "fabricTxId") {
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
		msg["fabricTxId"] = txid
		v, _ := json.Marshal(msg)
		invokeRequest.Value = string(v)
		tempData, _ := json.Marshal(invokeRequest)
		//logger.Debugf("SetArgsTxid msg is %s", tempData)
		(*args)[1] = string(tempData)
	}
}

func (sdk *sdkHandler) CheckBlockSyncState() {
	latestBlock, err := sdk.GetBlockHeight("")
	if err != nil {
		panic(fmt.Errorf("getblockHeight failed! err= %s", err.Error()))
	}
	logger.Warningf("the Current block Number: %d", latestBlock-1)
	sdk.SetCurrentBlockNumber(latestBlock - 1)
	checkTime := 20 * time.Second
	failedTimes := 0
	ticker := time.NewTicker(checkTime)
	for {
		select {
		case <-ticker.C:
			if failedTimes >= 3 {
				panic(fmt.Errorf("CheckBlockSyncState failed times is: 3"))
			}
			recordBlockNum := waitTxstatus.GlobalBlockNumber.Get()
			curChainBlockHeight, err := sdk.GetBlockHeight("")
			if err != nil {
				failedTimes++
				logger.Warningf("sdk GetBlockHeight err: %s", err.Error())
				continue
			}
			curChainBlockNum := curChainBlockHeight - 1
			if recordBlockNum != curChainBlockNum {
				logger.Warningf("It's different curChainBlockNum=%d, recordBlockNum=%d", curChainBlockNum, recordBlockNum)
				failedTimes++
			} else {
				//logger.Debugf("It's already sync block. curChainBlockNum=%d, recordBlockNum=%d", curChainBlockNum, recordBlockNum)
				failedTimes = 0
				continue
			}

			logger.Warningf("CheckBlockSyncState failed for %d times.", failedTimes)
			if failedTimes >= 3 {
				panic(fmt.Errorf("CheckBlockSyncState failed times is: 3"))
			}
		}
	}
}

func (sdk *sdkHandler) SetCurrentBlockNumber(num uint64) {
	waitTxstatus.GlobalBlockNumber.Put(num)
}

func (sdk *sdkHandler) JudgeFilterEventConnect(fRamBlock FuncGetRAMBLock, fLitsen FuncListenFilterBlock, ch chan<- EventBlockResponse) {
	interval := 5
	OldRamBlockNum, err := fRamBlock()
	if err != nil {
		panic("-JudgeFilterEventConnect-fRamBlock-err" + err.Error())
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	failedTimes := 0
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			curChainHeight, err := sdk.GetBlockHeight("")
			if err != nil {
				logger.Errorf("-JudgeFilterEventConnect-GetBlockHeight-err %s", err.Error())
			} else {
				NewRamBlockNum, err := fRamBlock()
				if err != nil {
					logger.Errorf("-JudgeFilterEventConnect-fRamBlock-err %s", err.Error())
					continue
				}
				//本地内存区块高度不变同时链上区块高度大于本地区块高度
				if OldRamBlockNum == NewRamBlockNum && curChainHeight > NewRamBlockNum+1 {
					logger.Errorf("--zyf-curBlockHeight=%d old=%d new=%d--\n", curChainHeight, OldRamBlockNum, NewRamBlockNum)
					failedTimes++
					logger.Errorf("-JudgeFilterEventConnect-block num diff times %d", failedTimes)
					if failedTimes == 2 {
						failedTimes = 0
						sdk.client.Event.DisConnect()
						logger.Errorf("--zyf-curBlockHeight=%d old=%d new=%d--\n", curChainHeight, OldRamBlockNum, NewRamBlockNum)
						if err := fLitsen("", int(NewRamBlockNum), ch); err != nil {
							logger.Errorf("-JudgeFilterEventConnect-reconnect err %s", err.Error())
						}
					}
				} else {
					OldRamBlockNum = NewRamBlockNum
				}
			}
		}
	}
}
func (sdk *sdkHandler) JudgeFullEventConnect(fRamBlock FuncGetRAMBLock, fLitsen FuncListenFullBlock, ch chan<- parseBlock.Block) {
	interval := 5
	OldRamBlockNum, err := fRamBlock()
	if err != nil {
		panic("-JudgeFullEventConnect-fRamBlock-err" + err.Error())
	}

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	failedTimes := 0
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			curChainHeight, err := sdk.GetBlockHeight("")
			if err != nil {
				logger.Errorf("-JudgeFullEventConnect-GetBlockHeight-err %s", err.Error())
			} else {
				NewRamBlockNum, err := fRamBlock()
				if err != nil {
					logger.Errorf("-JudgeFullEventConnect-fRamBlock-err %s", err.Error())
					continue
				}
				//本地内存区块高度不变同时链上区块高度大于本地区块高度
				if OldRamBlockNum == NewRamBlockNum && curChainHeight > NewRamBlockNum+1 {
					failedTimes++
					logger.Errorf("-JudgeFullEventConnect-block num diff times %d", failedTimes)
					if failedTimes == 2 {
						failedTimes = 0
						sdk.client.Event.DisConnect()
						if err := fLitsen("", int(NewRamBlockNum), ch); err != nil {
							logger.Errorf("-JudgeFullEventConnect-reconnect err %s", err.Error())
						}
					}
				} else {
					OldRamBlockNum = NewRamBlockNum
				}
			}
		}
	}
}

func (sdk *sdkHandler) EventFilterBlock(channelName string, startNum int, ch chan<- EventBlockResponse) error {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFilteredBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
	if err != nil {
		cancel()
		return err
	}
	return nil
}

func (sdk *sdkHandler) EventFullBlock(channelName string, startNum int, ch chan<- parseBlock.Block) error {
	if len(channelName) == 0 {
		channelName = sdk.client.Channel.ChannelId
	}

	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.ListenForFullBlock(ctx, *sdk.identity, startNum, eventName, channelName, ch)
	if err != nil {
		cancel()
		return err
	}

	return nil
}
