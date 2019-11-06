/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Orderer expose API's to communicate with orderers.
type Orderer struct {
	Name   string
	Uri    string
	Opts   []grpc.DialOption
	caPath string
	con    *grpc.ClientConn
	client orderer.AtomicBroadcastClient
}

const timeout = 5

// Broadcast Broadcast envelope to orderer for execution.
func (o *Orderer) Broadcast(envelope *common.Envelope) (*orderer.BroadcastResponse, error) {
	bcc, err := o.client.Broadcast(context.Background())
	if err != nil {
		return nil, err
	}
	defer bcc.CloseSend()
	bcc.Send(envelope)
	response, err := bcc.Recv()
	if err != nil {
		return nil, err
	}
	if response.Status != common.Status_SUCCESS {
		return nil, fmt.Errorf("unexpected status: %v", response.Status)
	}

	return response, err
}

// Deliver delivers envelope to orderer. Please note that new connection will be created on every call of Deliver.
func (o *Orderer) Deliver(envelope *common.Envelope) (*common.Block, error) {

	connection, err := grpc.Dial(o.Uri, o.Opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to orderer: %s err is: %v", o.Name, err)
	}
	defer connection.Close()

	dk, err := orderer.NewAtomicBroadcastClient(connection).Deliver(context.Background())
	if err != nil {
		return nil, err
	}
	if err := dk.Send(envelope); err != nil {
		return nil, err
	}
	var block *common.Block
	timer := time.NewTimer(time.Second * time.Duration(timeout))
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return nil, ErrOrdererTimeout
		default:
			response, err := dk.Recv()
			if err != nil {
				return nil, err
			}
			switch t := response.Type.(type) {
			case *orderer.DeliverResponse_Status:
				if t.Status == common.Status_SUCCESS {
					return block, nil
				} else {
					return nil, fmt.Errorf("orderer response with status: %v", t.Status)
				}
			case *orderer.DeliverResponse_Block:
				block = response.GetBlock()

			default:
				return nil, fmt.Errorf("unknown response type from orderer: %s", t)
			}
		}
	}
}

func (o *Orderer) getGenesisBlock(identity Identity, crypto CryptoSuite, channelId string) (*common.Block, error) {
	seekInfo := &orderer.SeekInfo{
		Start:    &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: 0}}},
		Stop:     &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: 0}}},
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}
	seekInfoBytes, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, err
	}

	creator, err := marshalProtoIdentity(identity)
	if err != nil {
		return nil, err
	}
	txId, err := newTransactionId(creator, crypto)
	if err != nil {
		return nil, err
	}

	headerBytes, err := channelHeader(common.HeaderType_DELIVER_SEEK_INFO, txId, channelId, 0, nil)
	signatureHeaderBytes, err := signatureHeader(creator, txId)
	if err != nil {
		return nil, err
	}
	header := header(signatureHeaderBytes, headerBytes)
	payloadBytes, err := payload(header, seekInfoBytes)
	if err != nil {
		return nil, err
	}
	payloadSignedBytes, err := crypto.Sign(payloadBytes, identity.PrivateKey)
	if err != nil {
		return nil, err
	}
	env := &common.Envelope{Payload: payloadBytes, Signature: payloadSignedBytes}
	return o.Deliver(env)
}

// newOrdererFromConfig create new Orderer from config
func newOrdererFromConfig(conf OrdererConfig) (*Orderer, error) {
	o := Orderer{Uri: conf.Host, caPath: conf.TlsPath}
	if !conf.UseTLS {
		o.Opts = []grpc.DialOption{grpc.WithInsecure()}
	} else {
		//creds, err := credentials.NewClientTLSFromFile(o.caPath, conf.DomainName)
		//if err != nil {
		//	return nil, fmt.Errorf("cannot read orderer %s credentials err is: %v", o.Name, err)
		//}
		certPool := x509.NewCertPool()
		if isCorrect := certPool.AppendCertsFromPEM(conf.TLSInfo[0]); !isCorrect {
			logger.Errorf("append certs failed for %s", conf.Host)
			return nil, errors.New("append certs failed")
		}

		creds := credentials.NewClientTLSFromCert(certPool, "")

		o.Opts = append(o.Opts, grpc.WithTransportCredentials(creds))
	}
	o.Opts = append(o.Opts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(1) * time.Minute,
			Timeout:             time.Duration(20) * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize)),
		grpc.WithTimeout(time.Second*30),
		grpc.WithBackoffConfig(grpc.BackoffConfig{
			MaxDelay: time.Second * 10,
		}))
	c, err := grpc.Dial(o.Uri, o.Opts...)
	if err != nil {
		return nil, err
	}
	o.con = c
	o.client = orderer.NewAtomicBroadcastClient(o.con)
	return &o, nil
}
