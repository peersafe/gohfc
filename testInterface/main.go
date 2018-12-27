package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc"
	"time"
)

var (
	logger   = logging.MustGetLogger("sdk")
	funcName = flag.String("function", "", "invoke,query,listen,checkordconn")
)

func main() {
	flag.Parse()
	handler, err := gohfc.InitSdkByFile("./zyfclient.yaml")
	if err != nil {
		logger.Error(err)
		return
	}

	//mm := make(chan int)
	//<-mm
	logger.Debugf("--testInterface main--")

	switch *funcName {
	case "invoke":
		res, err := handler.Invoke([]string{"invoke", "a", "b", "1"}, "mychannel", "factor")
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debugf("----invoke--TxID--%s\n", res.TxID)
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
