package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/peersafe/gohfc"
	"github.com/spf13/viper"
	"github.com/op/go-logging"
	"os"
)

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:2006-01-02 15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "gohfc")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("./examples invoke a b 1")
		fmt.Println("./examples query a")
		fmt.Println("./examples listen")
		return
	}
	if err := gohfc.InitSDK("./client.yaml"); err != nil {
		fmt.Println(err)
		return
	}

	peers := []string{"peer0"}
	if args[0] == "invoke" {
		result, err := gohfc.GetHandler().Invoke(args, peers, "orderer0")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	} else if args[0] == "query" {
		result, err := gohfc.GetHandler().Query(args, peers)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(result[0].Response.Response.GetPayload()))
	} else if args[0] == "listen" {
		ch, err := gohfc.GetHandler().ListenEvent("peer0", viper.GetString("other.localMspId"))
		if err != nil {
			fmt.Println(err)
			return
		}
		for {
			select {
			case block := <-ch:
				data, _ := json.Marshal(block.Error)
				fmt.Printf("%s\n", data)
			}
		}
	} else if args[0] == "queryQscc" {
		result, err := gohfc.GetHandler().QueryByQscc(args, peers)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(result[0].Response.Response.GetPayload()))
	} else {
		fmt.Println("----------args[0] err----------")
	}
}
