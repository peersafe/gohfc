package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc"
	"time"
)

var (
	logger   = logging.MustGetLogger("testInterface")
	funcName = flag.String("function", "", "invoke,query,listen,checkordconn")
)

func main() {
	flag.Parse()
	err := gohfc.InitSDK("./client.yaml")
	if err != nil {
		logger.Error(err)
		return
	}
	if gohfc.SetLogLevel(gohfc.GetConfigLogLevel(), "testInterface"); err != nil {
		logger.Error(err)
		return
	}
	logger.Debugf("--testInterface main--")

	switch *funcName {
	case "invoke":
		res, err := gohfc.GetHandler().Invoke([]string{"invoke", "a", "b", "1"}, "", "")
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debugf("----invoke--TxID--%s\n", res.TxID)
	case "query":
		resVal, err := gohfc.GetHandler().Query([]string{"query", "a"}, "", "")
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
		ch, err := gohfc.GetHandler().ListenEventFullBlock("", 3)
		if err != nil {
			logger.Error(err)
			return
		}
		for {
			select {
			case b := <-ch:
				logger.Debugf("------listen block num---%d\n", b.Header.Number)
			}
		}
	case "checkordconn":
		for {
			ok, err := gohfc.GetHandler().GetOrdererConnect()
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
