package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/peersafe/gohfc"
	"github.com/spf13/viper"
	"github.com/op/go-logging"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("./examples invoke a b 1")
		fmt.Println("./examples query a")
		fmt.Println("./examples listen")
		return
	}
	logging.SetLevel(logging.DEBUG,"grpc")
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
