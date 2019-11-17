/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/peersafe/gohfc/parseBlock"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

// FabricClient expose API's to work with Hyperledger Fabric
type FabricClient struct {
	Crypto   CryptoSuite
	Peers    map[string]map[string][]*Peer
	Orderers map[string][]*Orderer
	LocalConfig
}

// EventClient expose API's to work with listening Fabric
type EventClient struct {
	Crypto     CryptoSuite
	EventPeers map[string]*Peer
	LocalConfig
	Event *EventListener
}

// TODO: support multi channel
var chConfig *discovery.ConfigResult

// endorser groups for channel with chaincode
// the first map's key is the channel name add chaincode name, and the value is the endorsement policy, which is also the map struct
// the seconde map's key is the number of the endorsement policy, and the value is the detailed orgs' name, which is the slice struct
//                            |---0 [org1]
//                            |---1 [org2]     org1 or org2 or org3 for cc1
//              |----cc1------|---2 [org3]
// channel------|
//              |----cc2------|---0 [org1 & org2]   (org1 and org2) or org3 for cc2
//                            |---1 [org3]
var endorserGroups map[string]map[int][]string

func init() {
	endorserGroups = make(map[string]map[int][]string)
}

// CreateUpdateChannel read channel config generated (usually) from configtxgen and send it to orderer
// This step is needed before any peer is able to join the channel and before any future updates of the channel.
func (c *FabricClient) CreateUpdateChannel(identity Identity, path string, channelId string, orderer string) error {

	ord, ok := c.Orderers[channelId]
	if !ok {
		return ErrInvalidOrdererName
	}

	envelope, err := decodeChannelFromFs(path)
	if err != nil {
		return err
	}
	ou, err := buildAndSignChannelConfig(identity, envelope.GetPayload(), c.Crypto, channelId)
	if err != nil {
		return err
	}
	replay, err := ord[0].Broadcast(ou)
	if err != nil {
		return err
	}
	if replay.GetStatus() != common.Status_SUCCESS {
		return errors.New("error creating new channel. See orderer logs for more details")
	}
	return nil
}

func (c *FabricClient) ConfigUpdate(identity Identity, data []byte, channelId string, orderer string) error {
	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err := proto.Unmarshal(data, configUpdateEnvelope)
	if err != nil {
		return err

	}

	ord, ok := c.Orderers[channelId]
	if !ok {
		return ErrInvalidOrdererName
	}

	ou, err := buildAndSignConfigUpdate(identity, configUpdateEnvelope, c.Crypto, channelId)
	if err != nil {
		return err
	}
	replay, err := ord[0].Broadcast(ou)
	if err != nil {
		return err
	}
	if replay.GetStatus() != common.Status_SUCCESS {
		return errors.New("error creating new channel. See orderer logs for more details")
	}
	return nil
}

// JoinChannel send transaction to one or many Peers to join particular channel.
// Channel must be created before this operation using `CreateUpdateChannel` or manually using CLI interface.
// Orderers must be aware of this channel, otherwise operation will fail.
func (c *FabricClient) JoinChannel(identity Identity, channelId string, peers []string, orderer string) ([]*PeerResponse, error) {
	ord, ok := c.Orderers[channelId]
	if !ok {
		return nil, ErrInvalidOrdererName
	}

	execPeers := c.getPeers(channelId, peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}

	block, err := ord[0].getGenesisBlock(identity, c.Crypto, channelId)

	if err != nil {
		return nil, err
	}

	blockBytes, err := proto.Marshal(block)
	if err != nil {
		return nil, err
	}

	chainCode := ChainCode{Name: CSCC,
		Type:     ChaincodeSpec_GOLANG,
		Args:     []string{"JoinChain"},
		ArgBytes: blockBytes}

	invocationBytes, err := chainCodeInvocationSpec(chainCode)
	if err != nil {
		return nil, err
	}
	creator, err := marshalProtoIdentity(identity)
	if err != nil {
		return nil, err
	}
	txId, err := newTransactionId(creator, c.Crypto)
	if err != nil {
		return nil, err
	}
	ext := &peer.ChaincodeHeaderExtension{ChaincodeId: &peer.ChaincodeID{Name: CSCC}}
	channelHeaderBytes, err := channelHeader(common.HeaderType_ENDORSER_TRANSACTION, txId, "", 0, ext)
	if err != nil {
		return nil, err
	}

	sigHeaderBytes, err := signatureHeader(creator, txId)
	if err != nil {
		return nil, err
	}

	header := header(sigHeaderBytes, channelHeaderBytes)
	headerBytes, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}
	chainCodePropPl := new(peer.ChaincodeProposalPayload)
	chainCodePropPl.Input = invocationBytes

	chainCodePropPlBytes, err := proto.Marshal(chainCodePropPl)
	if err != nil {
		return nil, err
	}

	proposalBytes, err := proposal(headerBytes, chainCodePropPlBytes)
	if err != nil {
		return nil, err
	}

	proposal, err := signedProposal(proposalBytes, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	return sendToPeers(execPeers, proposal), nil
}

// InstallChainCode install chainCode to one or many peers. Peer must be in the channel where chaincode will be installed.
func (c *FabricClient) InstallChainCode(identity Identity, req *InstallRequest, peers []string) ([]*PeerResponse, error) {
	execPeers := c.getPeers("", peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}
	prop, err := createInstallProposal(identity, req, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	return sendToPeers(execPeers, proposal), nil

}

// InstantiateChainCode run installed chainCode to particular peer in particular channel.
// Chaincode must be installed using InstallChainCode or CLI interface before this operation.
// If this is first time running the chaincode operation must be `deploy`
// If this operation update existing chaincode operation must be `upgrade`
// collectionsConfig is configuration for private collections in versions >= 1.1. If not provided no private collections
// will be created. collectionsConfig can be specified when chaincode is upgraded.
func (c *FabricClient) InstantiateChainCode(identity Identity, req *ChainCode, peers []string, orderer string,
	operation string, collectionsConfig []CollectionConfig) (*orderer.BroadcastResponse, error) {
	ord, ok := c.Orderers[""]
	if !ok {
		return nil, ErrInvalidOrdererName
	}

	execPeers := c.getPeers("", peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}
	var collConfigBytes []byte
	if len(collectionsConfig) > 0 {
		collectionPolicy, err := collectionConfigToPolicy(collectionsConfig)
		if err != nil {
			return nil, err
		}
		collConfigBytes, err = proto.Marshal(&common.CollectionConfigPackage{Config: collectionPolicy})
		if err != nil {
			return nil, err
		}
	}

	prop, err := createInstantiateProposal(identity, req, operation, collConfigBytes, c.Crypto)
	if err != nil {
		return nil, err
	}

	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}

	transaction, err := createTransaction(prop.proposal, sendToPeers(execPeers, proposal))
	if err != nil {
		return nil, err
	}

	signedTransaction, err := c.Crypto.Sign(transaction, identity.PrivateKey)
	if err != nil {
		return nil, err
	}

	reply, err := ord[0].Broadcast(&common.Envelope{Payload: transaction, Signature: signedTransaction})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// QueryInstalledChainCodes get all chainCodes that are installed but not instantiated in one or many peers
func (c *FabricClient) QueryInstalledChainCodes(identity Identity, peers []string) ([]*ChainCodesResponse, error) {
	execPeers := c.getPeers("", peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}
	if len(identity.MspId) == 0 {
		return nil, ErrMspMissing
	}
	chainCode := ChainCode{
		Name: LSCC,
		Type: ChaincodeSpec_GOLANG,
		Args: []string{"getinstalledchaincodes"},
	}

	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}

	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)

	response := make([]*ChainCodesResponse, len(r))
	for idx, p := range r {
		ic := ChainCodesResponse{PeerName: p.Name, Error: p.Err}
		if p.Err != nil {
			ic.Error = p.Err
		} else {
			dec, err := decodeChainCodeQueryResponse(p.Response.Response.GetPayload())
			if err != nil {
				ic.Error = err
			}
			ic.ChainCodes = dec
		}
		response[idx] = &ic
	}
	return response, nil
}

// QueryInstantiatedChainCodes get all chainCodes that are running (instantiated) "inside" particular channel in peer
func (c *FabricClient) QueryInstantiatedChainCodes(identity Identity, channelId string, peers []string) ([]*ChainCodesResponse, error) {
	execPeers := c.getPeers(channelId, peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}

	prop, err := createTransactionProposal(identity, ChainCode{
		ChannelId: channelId,
		Name:      LSCC,
		Type:      ChaincodeSpec_GOLANG,
		Args:      []string{"getchaincodes"},
	}, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)
	response := make([]*ChainCodesResponse, len(r))
	for idx, p := range r {
		ic := ChainCodesResponse{PeerName: p.Name, Error: p.Err}
		if p.Err != nil {
			ic.Error = p.Err
		} else {
			dec, err := decodeChainCodeQueryResponse(p.Response.Response.GetPayload())
			if err != nil {
				ic.Error = err
			}
			ic.ChainCodes = dec
		}
		response[idx] = &ic
	}
	return response, nil
}

// QueryChannels returns a list of channels that peer/s has joined
func (c *FabricClient) QueryChannels(identity Identity, peers []string) ([]*QueryChannelsResponse, error) {
	execPeers := c.getPeers("", peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}

	chainCode := ChainCode{
		Name: CSCC,
		Type: ChaincodeSpec_GOLANG,
		Args: []string{"GetChannels"},
	}

	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)
	response := make([]*QueryChannelsResponse, 0, len(r))
	for _, pr := range r {
		peerResponse := QueryChannelsResponse{PeerName: pr.Name}
		if pr.Err != nil {
			peerResponse.Error = err
		} else {
			channels := new(peer.ChannelQueryResponse)
			if err := proto.Unmarshal(pr.Response.Response.Payload, channels); err != nil {
				peerResponse.Error = err

			} else {
				peerResponse.Channels = make([]string, 0, len(channels.Channels))
				for _, ci := range channels.Channels {
					peerResponse.Channels = append(peerResponse.Channels, ci.ChannelId)
				}
			}
		}
		response = append(response, &peerResponse)
	}
	return response, nil
}

// QueryChannelInfo get current block height, current hash and prev hash about particular channel in peer/s
func (c *FabricClient) QueryChannelInfo(identity Identity, channelId string, peers []string) ([]*QueryChannelInfoResponse, error) {
	execPeers := c.getPeers(channelId, peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}
	chainCode := ChainCode{
		ChannelId: channelId,
		Name:      QSCC,
		Type:      ChaincodeSpec_GOLANG,
		Args:      []string{"GetChainInfo", channelId},
	}

	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)

	response := make([]*QueryChannelInfoResponse, 0, len(r))
	for _, pr := range r {
		peerResponse := QueryChannelInfoResponse{PeerName: pr.Name}
		if pr.Err != nil {
			peerResponse.Error = err
		} else {
			bci := new(common.BlockchainInfo)
			if err := proto.Unmarshal(pr.Response.Response.Payload, bci); err != nil {
				peerResponse.Error = err

			} else {
				peerResponse.Info = bci
			}
		}
		response = append(response, &peerResponse)
	}
	return response, nil

}

// Query execute chainCode to one or many peers and return there responses without sending
// them to orderer for transaction - ReadOnly operation.
// Because is expected all peers to be in same state this function allows very easy horizontal scaling by
// distributing query operations between peers.
func (c *FabricClient) Query(identity Identity, chainCode ChainCode) ([]*QueryResponse, error) {
	//execPeers := c.getPeers(peers)
	//if len(peers) != len(execPeers) {
	//	return nil, ErrPeerNameNotFound
	//}
	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}

	//r := sendToPeers(execPeers, proposal)
	r, err := sendToOneEndorserPeer(proposal, chainCode.ChannelId, chainCode.Name)
	if err != nil {
		return nil, err
	}

	response := make([]*QueryResponse, 1)
	ic := QueryResponse{PeerName: r.Name, Error: r.Err}
	if r.Err != nil {
		ic.Error = r.Err
	} else {
		ic.Response = r.Response
	}
	response[0] = &ic

	return response, nil
}

func (c *EventClient) QueryByEvent(identity Identity, chainCode ChainCode) ([]*QueryResponse, error) {
	execPeers := c.getEventPeers()
	if len(execPeers) == 0 {
		return nil, ErrPeerNameNotFound
	}
	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)
	response := make([]*QueryResponse, len(r))
	for idx, p := range r {
		ic := QueryResponse{PeerName: p.Name, Error: p.Err}
		if p.Err != nil {
			ic.Error = p.Err
		} else {
			ic.Response = p.Response
		}
		response[idx] = &ic
	}
	return response, nil
}

// Invoke execute chainCode for ledger update. Peers that simulate the chainCode must be enough to satisfy the policy.
// When Invoke returns with success this is not granite that ledger was update. Invoke will return `transactionId`.
// This ID will be transactionId in events.
// It is responsibility of SDK user to build logic that handle successful and failed commits.
// If chaincode call `shim.Error` or simulation fails for other reasons this is considered as simulation failure.
// In such case Invoke will return the error and transaction will NOT be send to orderer. This transaction will NOT be
// committed to blockchain.
func (c *FabricClient) Invoke(identity Identity, chainCode ChainCode, peers []string, orderer string) (*InvokeResponse, error) {
	//ord, ok := c.Orderers[orderer]
	//if !ok {
	//	return nil, ErrInvalidOrdererName
	//}
	//execPeers := c.getPeers(peers)
	//if len(peers) != len(execPeers) {
	//	return nil, ErrPeerNameNotFound
	//}
	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	peerResponse, err := sendToEndorserGroups(proposal, chainCode.ChannelId, chainCode.Name)
	if err != nil {
		return nil, err
	}
	transaction, err := createTransaction(prop.proposal, peerResponse)
	if err != nil {
		return nil, err
	}
	signedTransaction, err := c.Crypto.Sign(transaction, identity.PrivateKey)
	if err != nil {
		return nil, err
	}
	/*
		reply, err := ord.Broadcast(&common.Envelope{Payload: transaction, Signature: signedTransaction})
		if err != nil {
			return nil, err
		}
	*/
	reply, err := c.ordererBroadcast(chainCode.ChannelId, &common.Envelope{Payload: transaction, Signature: signedTransaction})
	if err != nil {
		return nil, err
	}

	return &InvokeResponse{Status: reply.Status, TxID: prop.transactionId}, nil
}

func (c *FabricClient) ordererBroadcast(channelId string, envelope *common.Envelope) (*orderer.BroadcastResponse, error) {
	for key, orderer := range c.Orderers[channelId] {
		if reply, err := orderer.Broadcast(envelope); err == nil {
			return reply, nil
		} else {
			logger.Errorf("send to orderer %s failed!", key)
		}
	}

	return nil, errors.New("send to all orderers failed")
}

// QueryTransaction get data for particular transaction.
// TODO for now it only returns status of the transaction, and not the whole data (payload, endorsement etc)
func (c *FabricClient) QueryTransaction(identity Identity, channelId string, txId string, peers []string) ([]*QueryTransactionResponse, error) {
	execPeers := c.getPeers(channelId, peers)
	if len(peers) != len(execPeers) {
		return nil, ErrPeerNameNotFound
	}
	chainCode := ChainCode{
		ChannelId: channelId,
		Name:      QSCC,
		Type:      ChaincodeSpec_GOLANG,
		Args:      []string{"GetTransactionByID", channelId, txId}}

	prop, err := createTransactionProposal(identity, chainCode, c.Crypto)
	if err != nil {
		return nil, err
	}
	proposal, err := signedProposal(prop.proposal, identity, c.Crypto)
	if err != nil {
		return nil, err
	}
	r := sendToPeers(execPeers, proposal)

	response := make([]*QueryTransactionResponse, len(r))
	for idx, p := range r {
		qtr := QueryTransactionResponse{PeerName: p.Name, Error: p.Err}
		if p.Err != nil {
			qtr.Error = p.Err
		} else {
			dec, err := decodeTransaction(p.Response.Response.GetPayload())
			if err != nil {
				qtr.Error = err
			}
			qtr.StatusCode = dec
		}
		response[idx] = &qtr
	}
	return response, nil
}

// ListenForFullBlock will listen for events when new block is committed to blockchain and will return block height,
// list of all transactions in this block, there statuses and events associated with them.
// Listener is per channel, so user must create a new listener for every channel of interest.
// This event listener will start listen from newest block, and actual (raw) block data will NOT be returned.
// If user wants fo start listening from different blocks or want to receive full block bytes
// he/she must construct the listener manually and provide proper seek and block options.
// User must provide channel where events will be send and is responsibility for the user to read this channel.
// To cancel listening provide context with cancellation option and call cancel.
// User can listen for same events in same channel in multiple peers for redundancy using same `chan<- EventBlockResponse`
// In this case every peer will send its events, so identical events may appear more than once in channel.
func (c *EventClient) ListenForFullBlock(ctx context.Context, identity Identity, startNum int, channelId string, response chan<- parseBlock.Block) error {
	for _, eventpeer := range c.EventPeers {
		listener, err := newEventListener(ctx, c.Crypto, identity, *eventpeer, channelId, EventTypeFullBlock)
		if err != nil {
			logger.Warningf("create deliver client failed: %v", err)
			continue
		}
		if startNum < 0 {
			err = listener.SeekNewest()
		} else {
			err = listener.SeekRange(uint64(startNum), math.MaxUint64)
		}
		if err != nil {
			logger.Warningf("send envelope to peer failed: %v", err)
			continue
		}
		listener.Listen(response, nil)

		c.Event = listener
		return nil
	}

	return errors.New("connect all peers failed")

}

// ListenForFilteredBlock listen for events in blockchain. Difference with `ListenForFullBlock` is that event names
// will be returned but NOT events data. Also full block data will not be available.
// Other options are same as `ListenForFullBlock`.
func (c *EventClient) ListenForFilteredBlock(ctx context.Context, identity Identity, startNum int, eventPeer, channelId string, response chan<- EventBlockResponse) error {
	ep, ok := c.EventPeers[eventPeer]
	if !ok {
		return ErrPeerNameNotFound
	}
	listener, err := newEventListener(ctx, c.Crypto, identity, *ep, channelId, EventTypeFiltered)
	if err != nil {
		return err
	}
	if startNum < 0 {
		err = listener.SeekNewest()
	} else {
		err = listener.SeekRange(uint64(startNum), math.MaxUint64)
	}
	if err != nil {
		return err
	}
	listener.Listen(nil, response)

	c.Event = listener
	return nil
}

/*
// Listen v 1.0.4 -- port ==> 7053
func (c *FabricClient) Listen(ctx context.Context, identity *Identity, eventPeer, channelId, mspId string, response chan<- parseBlock.Block) error {
	ep, ok := c.EventPeers[eventPeer]
	if !ok {
		return ErrPeerNameNotFound
	}
	eventPort := &EventPort{
		event: EventListener{
			Context:   ctx,
			Peer:      *ep,
			Identity:  *identity,
			ChannelId: channelId,
			Crypto:    c.Crypto,
			FullBlock: false,
		},
	}
	c.EventPort = eventPort
	return c.EventPort.newEventListener(response, mspId)
}
*/
// NewFabricClientFromConfig create a new FabricClient from ClientConfig
func NewFabricClientFromConfig(config ClientConfig) (*FabricClient, error) {
	var crypto CryptoSuite
	var err error
	switch config.CryptoConfig.Family {
	case "ecdsa":
		crypto, err = NewECCryptSuiteFromConfig(config.CryptoConfig)
		if err != nil {
			return nil, err
		}
	case "gm":
		crypto, err = NewECCryptSuiteFromConfig(config.CryptoConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidAlgorithmFamily
	}

	client := FabricClient{Crypto: crypto, LocalConfig: config.LocalConfig}
	return &client, nil
}

// NewFabricClientFromConfig create a new EventClient from EventConfig
func NewEventClientFromConfig(config EventConfig) (*EventClient, error) {
	var crypto CryptoSuite
	var err error
	switch config.CryptoConfig.Family {
	case "ecdsa":
		crypto, err = NewECCryptSuiteFromConfig(config.CryptoConfig)
		if err != nil {
			return nil, err
		}
	case "gm":
		crypto, err = NewECCryptSuiteFromConfig(config.CryptoConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidAlgorithmFamily
	}

	client := EventClient{Crypto: crypto, LocalConfig: config.LocalConfig}
	return &client, nil
}

// NewFabricClient creates new client from provided config file.
func NewFabricClient(path string) (*FabricClient, error) {
	config, err := newClientConfig(path)
	if err != nil {
		return nil, err
	}
	return NewFabricClientFromConfig(*config)
}

func getOrderersFromChannelConfig(cr *discovery.ConfigResult) (map[string]OrdererConfig, error) {
	ocs := make(map[string]OrdererConfig)
	for key, value := range cr.Orderers {
		logger.Debugf("orderer org %s has %d orderers", key, len(value.Endpoint))
		// get the org's root tls
		tlsInfo := chConfig.Msps[key].TlsRootCerts
		for i, point := range value.Endpoint {
			logger.Debugf("%d point's host is %s and port is %d", i, point.Host, point.Port)
			// construct the orderer config struct
			var oc OrdererConfig
			oc.Host = fmt.Sprintf("%s:%d", point.Host, point.Port)
			oc.UseTLS = true
			oc.TLSInfo = tlsInfo
			orderName := fmt.Sprintf("%s-%d", key, i)
			ocs[orderName] = oc
		}
	}
	if len(ocs) <= 0 {
		return nil, errors.New("channel does not have any orderer")
	}
	return ocs, nil
}

func newOrdererHandle(clientConfig *ClientConfig) (map[string][]*Orderer, error) {
	var err error
	var sy = &sync.Mutex{}
	oHandles := make(map[string][]*Orderer)
	ordererChan := make(chan OrdererConfig)

	for channel := range clientConfig.CCofChannels {
		chConfig, err = discoveryChannelConfig(channel)
		if err != nil {
			return nil, err
		}
		ocs, err := getOrderersFromChannelConfig(chConfig)
		if err != nil {
			return nil, err
		}
		logger.Debugf("channel %s has %d orderers", channel, len(ocs))

		for name, o := range ocs {
			oHandle, err := newOrdererFromConfig(o)
			if err != nil {
				logger.Errorf("connect to orderer %s failed", o.Host)
				if clientConfig.IsReConnect {
					go ordererReConnect(clientConfig.ReConnTimeInterval, ordererChan, name, channel, sy)
					ordererChan <- o
				}
				continue
			}
			oHandle.Name = name

			oHandles[channel] = append(oHandles[channel], oHandle)
			logger.Debugf("make handle to orderer %s successful", name)

			if len(oHandles[channel]) == 0 {
				return nil, errors.New("no available orderer handle")
			}
		}
	}

	return oHandles, nil
}

// TODO: handle multi chaincodes
func getPeersFromDiscovery(channel string, chaincodes []string) (map[string][]ConnectionConfig, error) {
	egMsp := make(map[string]string)

	eps, err := DiscoveryEndorsePolicy(channel, chaincodes, nil)
	if err != nil {
		return nil, err
	}

	pConnConfigs := make(map[string][]ConnectionConfig)

	for _, ep := range eps {
		// the key is just a code name, which can be considered as the mspid, although their values are not equal
		for key, egs := range ep.EndorsersByGroups {
			logger.Debugf("org %s has %d available peers", key, len(egs))
			for _, eg := range egs {
				logger.Debugf("endorser msp id is %s and endpoint is %s", eg.MSPID, eg.Endpoint)
				egMsp[key] = eg.MSPID
				// get the peer's root tls
				tlsInfo := chConfig.Msps[eg.MSPID].TlsRootCerts
				// construct the connection config for the peer
				var cConfig ConnectionConfig
				cConfig.Host = eg.Endpoint
				//cConfig.OrgName = key
				cConfig.TLSInfo = tlsInfo
				cConfig.UseTLS = true
				cConfig.ChannelId = channel
				cConfig.MSPId = eg.MSPID
				// one key may have multi peers
				pConnConfigs[eg.MSPID] = append(pConnConfigs[eg.MSPID], cConfig)
			}
		}
		logger.Debugf("in channel %s chaincode %s has %d layouts", channel, ep.Chaincode, len(ep.Layouts))
		endorserGroup := make(map[int][]string)
		for k, layout := range ep.Layouts {
			for key, value := range layout.QuantitiesByGroup {
				logger.Debugf("the key is %s and value is %d in round %d", key, value, k)
				//orRulePeerNames = append(orRulePeerNames, key)
				endorserGroup[k] = append(endorserGroup[k], egMsp[key])
			}
		}
		logger.Debugf("in channel %s chaincode %s has %d endorsement group", channel, ep.Chaincode, len(endorserGroup))
		// channel name & chaincode name
		enGroupKey := channel + ep.Chaincode
		endorserGroups[enGroupKey] = endorserGroup
	}

	if 0 == len(pConnConfigs) {
		return nil, errors.New("channel does not have any available peer")
	}
	logger.Debugf("channel %s  have %d available peers", channel, len(pConnConfigs))

	return pConnConfigs, nil
}

func newPeerHandle(clientConfig *ClientConfig) (map[string]map[string][]*Peer, error) {
	pHandles := make(map[string]map[string][]*Peer)
	clientChan := make(chan ConnectionConfig)
	var sy = &sync.Mutex{}

	for channel, chaincodes := range clientConfig.CCofChannels {
		pHandles[channel] = make(map[string][]*Peer)
		pConnConfigs, err := getPeersFromDiscovery(channel, chaincodes)
		if err != nil {
			return nil, err
		}
		for key, o := range pConnConfigs {
			for _, p := range o {
				ph := pHandles[channel][key]
				if checkReconnect(p, ph) {
					continue
				}
				c, err := newConnection(&p)
				if err != nil {
					logger.Errorf("connect to peer %s failed", p.Host)
					if clientConfig.IsReConnect {
						go peerReConnect(clientConfig.ReConnTimeInterval, clientChan, sy)
						clientChan <- p
					}
					continue
				}
				var pHandle Peer
				pHandle.conn = c
				pHandle.client = newPeerFromConn(c)
				pHandle.Uri = p.Host
				pHandles[channel][key] = append(pHandles[channel][key], &pHandle)
				logger.Debugf("create handle to the org %s %s successful", key, p.Host)
			}
		}
		if len(pHandles[channel]) == 0 {
			return nil, errors.New("no available peer handle")
		}
	}
	close(clientChan)
	return pHandles, nil
}

func newEventPeerHandle(eventConfig *EventConfig) (map[string]*Peer, error) {
	pHandles := make(map[string]*Peer)

	for name, eventpeerConfig := range eventConfig.EventPeers {
		c, err := newConnection(&eventpeerConfig)
		if err != nil {
			logger.Errorf("connect to peer %s failed", eventpeerConfig.Host)
			return nil, err
		}
		var pHandle Peer
		pHandle.conn = c
		pHandle.client = newPeerFromConn(c)
		pHandles[name] = &pHandle
		logger.Debugf("create handle to the %s successful", eventpeerConfig.Host)
	}

	return pHandles, nil
}

func checkReconnect(config ConnectionConfig, peers []*Peer) bool {
	for _, peer := range peers {
		if peer.Uri == config.Host {
			return true
		}
	}
	return false
}

func (c FabricClient) getPeers(channel string, names []string) []*Peer {
	res := make([]*Peer, 0, len(names))
	for _, p := range names {
		if fp, ok := c.Peers[channel][p]; ok {
			res = append(res, fp[generateRangeNum(0, len(fp))])
		}
	}
	return res
}

func (c EventClient) getEventPeers() []*Peer {
	var res []*Peer
	for _, p := range c.EventPeers {
		res = append(res, p)
	}
	return res
}
