package main

import (
	"flag"
	"fmt"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc"
	"io/ioutil"
	"time"
	"encoding/asn1"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/golang/protobuf/proto"
)

var (
	logger   = logging.MustGetLogger("sdk")
	funcName = flag.String("function", "", "invoke,query,listen,checkordconn")
	num = flag.Uint64("num", 2, "invoke,query,listen,checkordconn")
)


type asn1Header struct {
	Number       int64
	PreviousHash []byte
	DataHash     []byte
}

func main() {
	flag.Parse()
	handler, err := gohfc.InitSdkByFile("./client.yaml")
	if err != nil {
		logger.Error(err)
		return
	}

	//mm := make(chan int)
	//<-mm
	logger.Debugf("--testInterface main--")

	switch *funcName {
	case "getblock":
		block, err := handler.GetBlockByNumber(*num,"mychannel")
		if err != nil {
			panic(err)
		}
		fmt.Println("getblock")
		fmt.Printf("num=%d\n",block.Header.Number)
		fmt.Printf("previoushash=%s\n",fmt.Sprintf("%x\n",block.Header.PreviousHash))
		fmt.Printf("dathash=%s\n",fmt.Sprintf("%x\n",block.Header.DataHash))
		asn1Header := asn1Header{
			PreviousHash: block.Header.PreviousHash,
			DataHash:     block.Header.DataHash,
			Number:     int64(block.Header.Number),
		}
		result, err := asn1.Marshal(asn1Header)
		if err != nil {
			panic(err)
		}
		fmt.Printf("cal %d num blockhash=%s\n",block.Header.Number,fmt.Sprintf("%x\n",util.ComputeSHA256(result)))
		args1 := []string{"GetBlockByHash","mychannel", string(block.Header.PreviousHash)}

		resps, err := handler.Query(args1,"mychannel","qscc")
		if err != nil {
			panic(err)
		} else if len(resps) == 0 {
			panic(err)
		}
		if resps[0].Error != nil {
			panic(err)
		}
		data := resps[0].Response.Response.Payload
		var blockA = new(common.Block)
		err = proto.Unmarshal(data, blockA)
		if err != nil {
			panic(err)
		}
		fmt.Printf("num=%d\n",blockA.Header.Number)
		fmt.Printf("previoushash=%s\n",fmt.Sprintf("%x\n",blockA.Header.PreviousHash))
		fmt.Printf("dathash=%s\n",fmt.Sprintf("%x\n",blockA.Header.DataHash))
	case "invoke":
		res, err := handler.Invoke([]string{"invoke", "a", "b", "1"}, "mychannel", "factor")
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debugf("----invoke--TxID--%s\n", res.TxID)
	case "parseBlock":
		byte, err := ioutil.ReadFile("./mychannelConfig.block")
		if err != nil {
			panic(err)
		}
		curBlock := utils.UnmarshalBlockOrPanic(byte)
		decodeBlock,err := handler.ParseCommonBlock(curBlock)
		if err != nil {
			panic(err)
		}
		//str, _ := json.Marshal(decodeBlock)
		logger.Debugf("----Decode Block----%v\n", decodeBlock.BlockType)
	case "queryBlock":
		resVal, err := handler.Query([]string{"GetConfigBlock", "mychannel"}, "mychannel", "cscc")
		if err != nil || len(resVal) == 0 {
			logger.Error(err)
			return
		}
		if resVal[0].Error != nil {
			logger.Error(resVal[0].Error)
			return
		}
		if resVal[0].Response.Response.GetStatus() != 200 {
			logger.Error(fmt.Errorf(resVal[0].Response.Response.GetMessage()))
			return
		}
		if err := ioutil.WriteFile("./mychannelConfig.block", resVal[0].Response.Response.GetPayload(), 0655); err != nil {
			logger.Error(err)
			return
		}
		logger.Debugf("----queryBlock--result--in mychannelConfig.block\n")
	case "query":
		resVal, err := handler.Query([]string{"query", "a"}, "mychannel", "factor")
		if err != nil || len(resVal) == 0 {
			logger.Error(err)
			return
		}
		if resVal[0].Error != nil {
			logger.Error(resVal[0].Error)
			return
		}
		if resVal[0].Response.Response.GetStatus() != 200 {
			logger.Error(fmt.Errorf(resVal[0].Response.Response.GetMessage()))
			return
		}
		logger.Debugf("----query--result--%s\n", resVal[0].Response.Response.GetPayload())
	case "listen":
		ch, err := handler.ListenEventFullBlock("", 0)
		if err != nil {
			logger.Error(err)
			return
		}
		for {
			select {
			case b := <-ch:
				logger.Debugf("------listen block num---%v\n", b)
			}
		}
	case "checkordconn":
		for {
			ok, err := handler.GetOrdererConnect()
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Debugf("the connect is %v", ok)
			time.Sleep(2 * time.Second)
		}
	default:
		flag.PrintDefaults()
	}
	return
}
