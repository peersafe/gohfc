package gohfc

import (
	"context"
	"fmt"
	"github.com/cendhu/fetch-block/src/events/parse"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
)

var sdklogger = logging.MustGetLogger("gohfc")

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:2006-01-02 15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "gohfc")
}

//sdk handler
type sdkHandler struct {
	client   *FabricClient
	identity *Identity
}

var handler sdkHandler

func InitSDK(configPath string) error {
	// initialize Fabric client
	var err error
	handler.client, err = NewFabricClient(configPath)
	if err != nil {
		sdklogger.Debugf("Error loading file %s err: %v", configPath, err)
		return err
	}
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		sdklogger.Debugf("Read file failed:", err.Error())
		return err
	}
	mspPath := viper.GetString("other.mspConfigPath")
	if mspPath == "" {
		return fmt.Errorf("yaml mspPath is empty")
	}
	findCert := func(path string) string {
		list, err := ioutil.ReadDir(path)
		if err != nil {
			sdklogger.Debug(err.Error())
			return ""
		}
		var file os.FileInfo
		for _, item := range list {
			if !item.IsDir() {
				if file == nil {
					file = item
				} else if item.ModTime().After(file.ModTime()) {
					file = item
				}
			}
		}
		return filepath.Join(path, file.Name())
	}
	prikey := findCert(filepath.Join(mspPath, "keystore"))
	pubkey := findCert(filepath.Join(mspPath, "signcerts"))
	if prikey == "" || pubkey == "" {
		return fmt.Errorf("prikey or cert is no such file")
	}
	sdklogger.Debugf("privateKey : %s", prikey)
	sdklogger.Debugf("publicKey : %s", pubkey)
	handler.identity, err = LoadCertFromFile(pubkey, prikey)
	if err != nil {
		sdklogger.Debugf("load cert from file failed:", err.Error())
		return err
	}

	return err
}

// GetHandler get sdk handler
func GetHandler() *sdkHandler {
	return &handler
}

// Invoke invoke cc
func (sdk *sdkHandler) Invoke(args []string, peers []string, ordername string) (*InvokeResponse, error) {
	chaincode, err := getChainCodeObj(args)
	if err != nil {
		return nil, err
	}
	return sdk.client.Invoke(sdk.identity, chaincode, peers, ordername)
}

// Query query cc
func (sdk *sdkHandler) Query(args []string, peers []string) ([]*QueryResponse, error) {
	chaincode, err := getChainCodeObj(args)
	if err != nil {
		return nil, err
	}

	return sdk.client.Query(sdk.identity, chaincode, peers)
}

func (sdk *sdkHandler) GetBlockHeight(peers []string) (uint64, error) {
	channelid := viper.GetString("other.channelId")
	if channelid == "" {
		return 0, fmt.Errorf("channelid  is empty")
	}
	args := []string{"GetChainInfo", channelid}
	sdklogger.Debugf("GetBlockHeight chainId %s", channelid)
	resps, err := sdk.QueryByQscc(args, peers)
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

// Query query qscc
func (sdk *sdkHandler) QueryByQscc(args []string, peers []string) ([]*QueryResponse, error) {
	channelid := viper.GetString("other.channelId")
	mspId := viper.GetString("other.localMspId")
	if channelid == "" || mspId == "" {
		return nil, fmt.Errorf("channelid or ccname or mspid is empty")
	}

	channel := &Channel{
		MspId:       mspId,
		ChannelName: channelid,
	}
	chaincode := ChainCode{
		Channel: channel,
		Type:    ChaincodeSpec_GOLANG,
		Name:    "qscc",
		Args:    args,
	}

	return sdk.client.Query(sdk.identity, &chaincode, peers)
}

func (sdk *sdkHandler) ListenEvent(peername, mspid string) (chan parse.Block, error) {
	if peername == "" || mspid == "" {
		return nil, fmt.Errorf("ListenEvent peername or mspid is empty ")
	}
	ch := make(chan parse.Block)
	ctx, cancel := context.WithCancel(context.Background())
	err := sdk.client.Listen(ctx, sdk.identity, peername, mspid, ch)
	if err != nil {
		cancel()
		return nil, err
	}
	return ch, nil
}

func getChainCodeObj(args []string) (*ChainCode, error) {
	channelid := viper.GetString("other.channelId")
	chaincodeName := viper.GetString("other.chaincodeName")
	chaincodeVersion := viper.GetString("other.chaincodeVersion")
	mspId := viper.GetString("other.localMspId")
	if channelid == "" || chaincodeName == "" || chaincodeVersion == "" || mspId == "" {
		return nil, fmt.Errorf("channelid or ccname or ccver  or mspId is empty")
	}
	channel := &Channel{
		MspId:       mspId,
		ChannelName: channelid,
	}

	chaincode := ChainCode{
		Channel: channel,
		Type:    ChaincodeSpec_GOLANG,
		Name:    chaincodeName,
		Version: chaincodeVersion,
		Args:    args,
	}
	return &chaincode, nil
}
