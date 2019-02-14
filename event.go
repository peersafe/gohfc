/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"context"
	"encoding/pem"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"io"
	"time"
	"fmt"
	"golang.org/x/crypto/sha3"
	"github.com/cendhu/fetch-block/src/events/parse"
)

const GRPC_MAX_SIZE = 100 * 1024 * 1024

// BlockEventResponse holds event response when block is committed to peer.
type BlockEventResponse struct {
	// Error is error message.
	Error error
	// TxId is transaction id that generates this event
	IsVaild          bool
	BlockHeight      int64
	BlockSize        int64
	BlockHash        string
	BlockPreHash     string
	TxTime 			 time.Time
	TxIndex          int
	TxSize 			 int
	TxID             string
	TxHash           string
	TxContent        []byte
	TxCount 		 int
	TxType 			 string
	BlockConfirmTime int64
	ChannelName      string
	ChainCodeName    string
	ChainCodeVersion string
	Status           int32
	ChainCodeInput   [][]byte
	CCEvents         []*CCEvent
	BlockTime 		 time.Time
}

// CCEvent represent custom event send from chaincode using `stub.SetEvent`
type CCEvent struct {
	EventName    string
	EventPayload []byte
}

type eventHub struct {
	connection *grpc.ClientConn
	client     peer.Events_ChatClient
}

func (e *eventHub) connect(ctx context.Context, p *Peer) error {
	p.Opts = append(p.Opts, grpc.WithBlock(), grpc.WithTimeout(5*time.Second),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(GRPC_MAX_SIZE),
			grpc.MaxCallSendMsgSize(GRPC_MAX_SIZE)))
	conn, err := grpc.Dial(p.Uri, p.Opts...)
	if err != nil {
		return err
	}
	e.connection = conn
	event := peer.NewEventsClient(conn)
	cl, err := event.Chat(ctx)
	if err != nil {
		return err
	}
	e.client = cl
	return nil
}

func (e *eventHub) register(mspId string, identity *Identity, crypto CryptoSuite) error {
	creator, err := proto.Marshal(&msp.SerializedIdentity{
		Mspid:   mspId,
		IdBytes: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: identity.Certificate.Raw})})
	if err != nil {
		return err
	}

	interest := &peer.Event{Event: &peer.Event_Register{Register: &peer.Register{
		Events: []*peer.Interest{
			{EventType: peer.EventType_BLOCK},
		}}}, Creator: creator}
	evtBytes, err := proto.Marshal(interest)
	if err != nil {
		return err
	}

	sb, err := crypto.Sign(evtBytes, identity.PrivateKey)
	if err != nil {
		return err
	}
	sigEvent := peer.SignedEvent{EventBytes: evtBytes, Signature: sb}
	if err = e.client.Send(&sigEvent); err != nil {
		return err
	}
	return nil
}

func (e *eventHub) disconnect() {
	e.client.CloseSend()
	e.connection.Close()
}

func newEventListener(ctx context.Context, response chan<- parse.Block, crypto CryptoSuite, identity *Identity, mspId string, p *Peer) error {
	hub := new(eventHub)
	err := hub.connect(ctx, p)
	if err != nil {
		return err
	}
	err = hub.register(mspId, identity, crypto)
	if err != nil {
		return err
	}
	go hub.readBlock(response)
	return nil
}

func (e *eventHub) readBlock(response chan<- parse.Block) {
	for {
		in, err := e.client.Recv()
		if err == io.EOF {
			e.disconnect()
			return
		}
		if err != nil {
			response <- parse.Block{Error: err}
			e.disconnect()
			return
		}

		switch in.Event.(type) {
		case *peer.Event_Block:
			size := uint64(len(in.String()))
			response <- parse.ParseBlock(in.GetBlock(), size)
		}
	}
}

func DecodeEventBlock(pl []byte, blockNum uint64, idx int, metadata [][]byte, blockHeader *common.BlockHeader, now time.Time, count, size int) BlockEventResponse {
	response := BlockEventResponse{}
	envelope := new(common.Envelope)
	payload := new(common.Payload)
	header := new(common.ChannelHeader)
	ex := &peer.ChaincodeHeaderExtension{}
	if err := proto.Unmarshal(pl, envelope); err != nil {
		response.Error = err
		return response
	}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		response.Error = err
	}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, header); err != nil {
		response.Error = err
		return response
	}
	if err := proto.Unmarshal(header.Extension, ex); err != nil {
		response.Error = err
		return response
	}

	h := sha3.New256()
	h.Write(pl)
	hash := h.Sum(nil)

	txsFltr := util.TxValidationFlags(metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	response.IsVaild = txsFltr.IsValid(idx)
	response.BlockHeight = int64(blockNum)
	response.BlockHash = fmt.Sprintf("%x", blockHeader.DataHash)
	response.BlockPreHash = fmt.Sprintf("%x", blockHeader.PreviousHash)
	response.BlockSize = int64(size)
	response.BlockTime = now
	response.TxTime = time.Unix(header.Timestamp.Seconds, int64(header.Timestamp.Nanos))
	response.TxIndex = idx
	response.TxCount = count
	response.TxID = header.TxId
	response.TxHash = fmt.Sprintf("%x", hash)
	response.TxSize = len(pl)
	response.TxType = common.HeaderType_name[header.Type]
	response.TxContent = pl
	response.BlockConfirmTime = (response.BlockTime.UnixNano() - response.TxTime.UnixNano()) / 1000 / 1000
	response.ChannelName = header.ChannelId
	if ex.ChaincodeId != nil {
		response.ChainCodeName = ex.ChaincodeId.Name
		response.ChainCodeVersion = ex.ChaincodeId.Version
	}
	response.Status = int32(metadata[2][idx])
	if common.HeaderType(header.Type) == common.HeaderType_ENDORSER_TRANSACTION {
		tx := &peer.Transaction{}
		err := proto.Unmarshal(payload.Data, tx)
		if err != nil {
			response.Error = err
			return response
		}

		chainCodeActionPayload := &peer.ChaincodeActionPayload{}

		err = proto.Unmarshal(tx.Actions[0].Payload, chainCodeActionPayload)
		if err != nil {
			response.Error = err
			return response
		}

		chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
		err = proto.Unmarshal(chainCodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
		if err != nil {
			response.Error = err
			return response
		}

		chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(chaincodeProposalPayload.Input, chaincodeInvocationSpec)
		if err != nil {
			response.Error = err
			return response
		}
		response.ChainCodeInput = chaincodeInvocationSpec.GetChaincodeSpec().GetInput().Args

		propRespPayload := &peer.ProposalResponsePayload{}
		err = proto.Unmarshal(chainCodeActionPayload.Action.ProposalResponsePayload, propRespPayload)
		if err != nil {
			response.Error = err
			return response
		}

		caPayload := &peer.ChaincodeAction{}
		err = proto.Unmarshal(propRespPayload.Extension, caPayload)
		if err != nil {
			response.Error = err
			return response
		}
		ccEvent := &peer.ChaincodeEvent{}
		err = proto.Unmarshal(caPayload.Events, ccEvent)
		if err != nil {
			response.Error = err
			return response
		}
		if ccEvent != nil {
			response.CCEvents = append(response.CCEvents, &CCEvent{EventName: ccEvent.EventName, EventPayload: ccEvent.Payload})
		}
	}
	return response
}
