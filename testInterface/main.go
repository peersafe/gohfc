package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc"
)

var (
	logger   = logging.MustGetLogger("testInterface")
	funcName = flag.String("function", "", "invoke,query,listenfull(7051),listen(7053),checkordconn")
)

func main() {
	flag.Parse()
	err := gohfc.InitSDK("./client.yaml")
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Debugf("--Interface main--")

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
	case "listenfull":
		ch, err := gohfc.GetHandler().ListenEventFullBlock("", 3)
		if err != nil {
			logger.Errorf("ListenEventFullBlock err = %s", err.Error())
			return
		}

		for {
			select {
			case b := <-ch:
				if b.Error != nil {
					logger.Errorf("ListenEventFullBlock err = %s", b.Error.Error())
					continue
				}
				logger.Debugf("------listen block num---%d\n", b.Header.Number)
				if len(b.Transactions) == 0 {
					logger.Debugf("ListenEventFullBlock Config Block BlockNumber= %d, ", b.Header.Number)
				} else {
					//aa,_ := json.Marshal(b)
					//logger.Debugf("---%s\n",aa)
				}
			}
		}
	default:
		flag.PrintDefaults()
	}
	return
}
