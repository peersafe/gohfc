package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/protos/peer"
)

func MakePackGolangCC(SrcPath, Namespace string) ([]byte, error) {

	twBuf := new(bytes.Buffer)
	tw := tar.NewWriter(twBuf)

	var gzBuf bytes.Buffer
	zw := gzip.NewWriter(&gzBuf)

	_, err := os.Stat(SrcPath)
	if err != nil {
		return nil, err
	}
	rootpath := filepath.Join(SrcPath, "src", Namespace)
	err = filepath.Walk(rootpath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Mode = 0100000
			header.Name = strings.TrimPrefix(path, SrcPath)
			fmt.Println("-------------header.Name--------", header.Name)
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)

			return err
		})
	if err != nil {
		tw.Close()
		return nil, err
	}

	_, err = zw.Write(twBuf.Bytes())
	if err != nil {
		return nil, err
	}
	tw.Close()
	zw.Close()

	return gzBuf.Bytes(), nil
}

func CreateUploadChaincodeArgs(ChainCodeName, ChainCodeVersion, SrcPath, Namespace string) ([]string, error) {

	var packageBytes []byte
	var err error

	packageBytes, err = MakePackGolangCC(SrcPath, Namespace)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	depSpec, err := proto.Marshal(&peer.ChaincodeDeploymentSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: ChainCodeName, Path: Namespace, Version: ChainCodeVersion},
			Type:        peer.ChaincodeSpec_GOLANG,
		},
		CodePackage:   packageBytes,
		EffectiveDate: &timestamp.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())},
	})
	return []string{"install", string(depSpec)}, err
}

func CreateDeplocyChaincodeArgs(operation, chainid, ccname, ccversion, policy, initargs string) ([]string, error) {
	input := peer.ChaincodeInput{}
	if err := json.Unmarshal([]byte(initargs), &input); err != nil {
		return nil, fmt.Errorf("deploy initargs error : %s\n", err.Error())
	}
	policyEnvelope, err := cauthdsl.FromString(policy)
	if err != nil {
		return nil, fmt.Errorf("policy FromString error : %s\n", err.Error())
	}
	marshPolicy, err := proto.Marshal(policyEnvelope)
	if err != nil {
		return nil, fmt.Errorf("policy proto.Marshal error : %s\n", err.Error())
	}
	depSpec, err := proto.Marshal(&peer.ChaincodeDeploymentSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: ccname, Version: ccversion},
			Type:        peer.ChaincodeSpec_GOLANG,
			Input:       &input,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal ChaincodeInput error : %s\n", err.Error())
	}
	args := []string{operation, chainid, string(depSpec), string(marshPolicy), "escc", "vscc"}
	return args, nil
}
