package gohfc

import (
	"context"
	"reflect"
	"strings"
	//"errors"
	"time"

	"github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric/protos/gossip"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/peersafe/gohfc/discoveryClient"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	//mSinger *cdiscovery.MemoizeSigner
	signer *discoveryClient.Sign

	discoveryClients []*discoveryClient.PClient
	tlsHashs         map[string][]byte
)

func NewDiscoveryClient(conf map[string]ConnectionConfig, channelConfig *ChannelConfig) error {
	signer = discoveryClient.NewSign(channelConfig.MspConfigPath, channelConfig.LocalMspId)
	if signer == nil {
		return errors.New("construct signer failed")
	}

	for key, value := range conf {
		grpcConn, err := NewConnection(&value)
		if err != nil {
			logger.Errorf("make grpc connection to %s failed: %s", key, err.Error())
			continue
		}
		dc := discoveryClient.NewPClient(value.Host, key, grpcConn, signer.Sign, 0)
		if nil == dc {
			logger.Errorf("make discovery client for %s failed", key)
			continue
		}
		discoveryClients = append(discoveryClients, dc)
	}

	if 0 == len(discoveryClients) {
		return errors.New("make discovery client for all discovery peers failed")
	}

	return nil
}

func DiscoveryChannelPeers(channel string) ([]*ChannelPeer, error) {
	if "" == channel {
		return nil, errors.New("the input channel is empty")
	}

	req := discoveryClient.NewRequest()
	req = req.OfChannel(channel)
	req = req.AddPeersQuery()
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	auth := &discovery.AuthInfo{
		ClientIdentity:    signer.Creator,
		ClientTlsCertHash: baseTLSCertInfo.certHash, //comm.TLSCertHash,
	}

	for _, client := range discoveryClients {
		response, err := client.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send discovery channel peers request to %s failed %s", client.Name, err.Error())
			continue
		}
		peers, err := response.ForChannel(channel).Peers()
		var peerSlices []*ChannelPeer
		for _, p := range peers {
			peerSlices = append(peerSlices, rawPeerToChannelPeer(p))
		}
		return peerSlices, nil
	}

	return nil, errors.New("send discovery channel peers request to all peers failed!")
}

func DiscoveryLocalPeers(channel string) ([]*LocalPeer, error) {
	if "" == channel {
		return nil, errors.New("the input channel is empty")
	}

	req := discoveryClient.NewRequest()
	req = req.AddLocalPeersQuery()

	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	auth := &discovery.AuthInfo{
		ClientIdentity:    signer.Creator,
		ClientTlsCertHash: baseTLSCertInfo.certHash, //comm.TLSCertHash,
	}

	for _, client := range discoveryClients {
		response, err := client.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send discovery channel peers request to %s failed %s", client.Name, err.Error())
			continue
		}
		peers, err := response.ForChannel(channel).Peers()
		var peerSlices []*LocalPeer
		for _, p := range peers {
			peerSlices = append(peerSlices, rawPeerToLocalPeer(p))
		}
		return peerSlices, nil
	}

	return nil, errors.New("send discovery local peers request to all peers failed!")
}

func DiscoveryChannelConfig(channel string) (*discovery.ConfigResult, error) {
	if "" == channel {
		return nil, errors.New("the input channel is empty")
	}

	req := discoveryClient.NewRequest().OfChannel(channel).AddConfigQuery()
	auth := &discovery.AuthInfo{
		ClientIdentity:    signer.Creator,
		ClientTlsCertHash: baseTLSCertInfo.certHash, //comm.TLSCertHash,
	}

	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

	for _, dc := range discoveryClients {
		response, err := dc.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		chanConf, err := response.ForChannel(channel).Config()
		return chanConf, nil
	}

	return nil, errors.New("send discovery channel config to all peers failed")
}

func DiscoveryEndorsePolicy(channel string, chaincodes *[]string, collections *map[string]string) ([]*endorsermentDescriptor, error) {
	var ci []*discovery.ChaincodeInterest

	ccAndCol := &chaincodesAndCollections{
		Chaincodes:  chaincodes,
		Collections: collections,
	}
	cc2collections, err := ccAndCol.parseInput()
	if err != nil {
		return nil, err
	}

	//var ccCalls []*discovery.ChaincodeCall

	for _, cc := range *ccAndCol.Chaincodes {
		ci = append(ci, &discovery.ChaincodeInterest{
			Chaincodes: []*discovery.ChaincodeCall{{
				Name: cc,
				CollectionNames: cc2collections[cc],
			}},
		})
		//ccCalls = append(ccCalls, &discovery.ChaincodeCall{
		//	Name:            cc,
		//	CollectionNames: cc2collections[cc],
		//})
	}

	//req, err := discoveryClient.NewRequest().OfChannel(channel).AddEndorsersQuery(&discovery.ChaincodeInterest{Chaincodes: ccCalls})
	req, err := discoveryClient.NewRequest().OfChannel(channel).AddEndorsersQuery(ci...)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating request")
	}
	auth := &discovery.AuthInfo{
		ClientIdentity:    signer.Creator,
		ClientTlsCertHash: baseTLSCertInfo.certHash, //comm.TLSCertHash,
	}

	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

	for _, dc := range discoveryClients {
		response, err := dc.RowSend(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		if len(response.Results) == 0 {
			return nil, errors.New("empty results")
		}

		if e := response.Results[0].GetError(); e != nil {
			return nil, errors.Errorf("server returned: %s", e.Content)
		}

		ccQueryRes := response.Results[0].GetCcQueryRes()
		if ccQueryRes == nil {
			return nil, errors.Errorf("server returned response of unexpected type: %v", reflect.TypeOf(response.Results[0]))
		}
		endorser := parseEndorsementDescriptors(ccQueryRes.Content)
		return endorser, nil
	}

	return nil, errors.New("empty results")
}

type dclient struct {
	clients []*discoveryClient.PClient
	signer  *discoveryClient.Sign
}

/*
func newDclient(fc *FabricClient) (*dclient, error) {
	if fc == nil {
		return nil, errors.New("fabric client is nil")
	}
	signer := discoveryClient.NewSign(fc.Channel.MspConfigPath, fc.Channel.LocalMspId)
	if signer == nil {
		return nil, errors.New("construct signer failed")
	}
	var c dclient
	for k, peer := range fc.Peers {
		logger.Debugf("create %s dc for %s", k, peer.Uri)
		if peer.conn != nil {
			dc := discoveryClient.NewPClient(peer.Uri, peer.Name, peer.conn, signer.Sign, 0)
			if dc == nil {
				logger.Errorf("construct discovery client for %s failed!", peer.Uri)
				continue
			}

			c.clients = append(c.clients, dc)
		}
	}

	c.signer = signer
	return &c, nil
}
*/
/*
func (c *dclient) discoveryChannelPeers(channel string) []*ChannelPeer{
	req := discoveryClient.NewRequest()
	req = req.OfChannel(channel)
	req = req.AddPeersQuery()

	for _, dc := range c.clients {
		auth := &discovery.AuthInfo{
			ClientIdentity: c.signer.Creator,
			ClientTlsCertHash: handler.client.Peers[dc.Name].tlsCertHash, //comm.TLSCertHash,
		}

		timeout, _ := context.WithTimeout(context.Background(), time.Second * 10)
		//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

		response, err := dc.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		peers, err := response.ForChannel(channel).Peers()
		var peerSlices []*ChannelPeer
		for _, p := range peers {
			peerSlices = append(peerSlices, rawPeerToChannelPeer(p))
		}
		return peerSlices
	}

	return nil
}

func (c *dclient) discoveryLocalPeers() []*LocalPeer {
	req := discoveryClient.NewRequest()
	req = req.AddLocalPeersQuery()

	for _, dc := range c.clients {
		auth := &discovery.AuthInfo{
			ClientIdentity: c.signer.Creator,
			ClientTlsCertHash: handler.client.Peers[dc.Name].tlsCertHash, //comm.TLSCertHash,
		}

		timeout, _ := context.WithTimeout(context.Background(), time.Second * 10)
		//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

		response, err := dc.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		peers, err := response.ForLocal().Peers()
		var peerSlices []*LocalPeer
		for _, p := range peers {
			peerSlices = append(peerSlices, rawPeerToLocalPeer(p))
		}
		return peerSlices
	}

	return nil
}

func (c *dclient) discoveryChannelConfig(channel string) *discovery.ConfigResult{
	req := discoveryClient.NewRequest().OfChannel(channel).AddConfigQuery()

	for _, dc := range c.clients {
		auth := &discovery.AuthInfo{
			ClientIdentity: c.signer.Creator,
			ClientTlsCertHash: handler.client.Peers[dc.Name].tlsCertHash, //comm.TLSCertHash,
		}

		timeout, _ := context.WithTimeout(context.Background(), time.Second * 10)
		//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

		response, err := dc.Send(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		chanConf, err := response.ForChannel(channel).Config()
		return chanConf
	}

	return nil
}

func (c *dclient) discoveryEndorsePolicy(channel string, chaincodes *[]string, collections *map[string]string) ([]*endorsermentDescriptor, error){
	ccAndCol := &chaincodesAndCollections{
		Chaincodes:  chaincodes,
		Collections: collections,
	}
	cc2collections, err := ccAndCol.parseInput()
	if err != nil {
		return nil, err
	}

	var ccCalls []*discovery.ChaincodeCall

	for _, cc := range *ccAndCol.Chaincodes {
		ccCalls = append(ccCalls, &discovery.ChaincodeCall{
			Name:            cc,
			CollectionNames: cc2collections[cc],
		})
	}

	req, err := discoveryClient.NewRequest().OfChannel(channel).AddEndorsersQuery(&discovery.ChaincodeInterest{Chaincodes: ccCalls})
	if err != nil {
		return nil, errors.Wrap(err, "failed creating request")
	}

	for _, dc := range c.clients {
		auth := &discovery.AuthInfo{
			ClientIdentity: c.signer.Creator,
			ClientTlsCertHash: handler.client.Peers[dc.Name].tlsCertHash, //comm.TLSCertHash,
		}

		timeout, _ := context.WithTimeout(context.Background(), time.Second * 10)
		//timeout, cancel := context.WithTimeout(context.Background(), time.Second * 10)

		response, err := dc.RowSend(timeout, req, auth)
		if err != nil {
			logger.Errorf("send failed %s", err.Error())
			continue
		}
		if len(response.Results) == 0 {
			return nil, errors.New("empty results")
		}

		if e := response.Results[0].GetError(); e != nil {
			return nil, errors.Errorf("server returned: %s", e.Content)
		}

		ccQueryRes := response.Results[0].GetCcQueryRes()
		if ccQueryRes == nil {
			return nil, errors.Errorf("server returned response of unexpected type: %v", reflect.TypeOf(response.Results[0]))
		}
		endorser := parseEndorsementDescriptors(ccQueryRes.Content)
		return endorser, nil
	}

	return nil, errors.New("empty results")
}

*/

type ChannelPeer struct {
	MSPID        string
	LedgerHeight uint64
	Endpoint     string
	Identity     string
	Chaincodes   []string
}

type LocalPeer struct {
	MSPID    string
	Endpoint string
	Identity string
}

func rawPeerToChannelPeer(p *discoveryClient.Peer) *ChannelPeer {
	var ledgerHeight uint64
	var ccs []string
	if p.StateInfoMessage != nil && p.StateInfoMessage.GetStateInfo() != nil && p.StateInfoMessage.GetStateInfo().Properties != nil {
		properties := p.StateInfoMessage.GetStateInfo().Properties
		ledgerHeight = properties.LedgerHeight
		for _, cc := range properties.Chaincodes {
			if cc == nil {
				continue
			}
			ccs = append(ccs, cc.Name)
		}
	}
	var endpoint string
	if p.AliveMessage != nil && p.AliveMessage.GetAliveMsg() != nil && p.AliveMessage.GetAliveMsg().Membership != nil {
		endpoint = p.AliveMessage.GetAliveMsg().Membership.Endpoint
	}
	sID := &msp.SerializedIdentity{}
	proto.Unmarshal(p.Identity, sID)
	return &ChannelPeer{
		MSPID:        p.MSPID,
		Endpoint:     endpoint,
		LedgerHeight: ledgerHeight,
		Identity:     string(sID.IdBytes),
		Chaincodes:   ccs,
	}
}

func rawPeerToLocalPeer(p *discoveryClient.Peer) *LocalPeer {
	var endpoint string
	if p.AliveMessage != nil && p.AliveMessage.GetAliveMsg() != nil && p.AliveMessage.GetAliveMsg().Membership != nil {
		endpoint = p.AliveMessage.GetAliveMsg().Membership.Endpoint
	}
	sID := &msp.SerializedIdentity{}
	proto.Unmarshal(p.Identity, sID)
	return &LocalPeer{
		MSPID:    p.MSPID,
		Endpoint: endpoint,
		Identity: string(sID.IdBytes),
	}
}

type chaincodesAndCollections struct {
	Chaincodes  *[]string
	Collections *map[string]string
}

func (ec *chaincodesAndCollections) existsInChaincodes(chaincodeName string) bool {
	for _, cc := range *ec.Chaincodes {
		if chaincodeName == cc {
			return true
		}
	}
	return false
}

func (ec *chaincodesAndCollections) parseInput() (map[string][]string, error) {
	var emptyChaincodes []string
	if ec.Chaincodes == nil {
		ec.Chaincodes = &emptyChaincodes
	}
	var emptyCollections map[string]string
	if ec.Collections == nil {
		ec.Collections = &emptyCollections
	}

	res := make(map[string][]string)

	for _, cc := range *ec.Chaincodes {
		res[cc] = nil
	}

	for cc, collections := range *ec.Collections {
		if !ec.existsInChaincodes(cc) {
			return nil, errors.Errorf("a collection specified chaincode %s but it wasn't specified with a chaincode flag", cc)
		}
		res[cc] = strings.Split(collections, ",")
	}
	return res, nil
}

type endorser struct {
	MSPID        string
	LedgerHeight uint64
	Endpoint     string
	Identity     string
}

type endorsermentDescriptor struct {
	Chaincode         string
	EndorsersByGroups map[string][]endorser
	Layouts           []*discovery.Layout
}

func parseEndorsementDescriptors(descriptors []*discovery.EndorsementDescriptor) []*endorsermentDescriptor {
	var res []*endorsermentDescriptor
	for _, desc := range descriptors {
		endorsersByGroups := make(map[string][]endorser)
		for grp, endorsers := range desc.EndorsersByGroups {
			for _, p := range endorsers.Peers {
				endorsersByGroups[grp] = append(endorsersByGroups[grp], endorserFromRaw(p))
			}
		}
		res = append(res, &endorsermentDescriptor{
			Chaincode:         desc.Chaincode,
			Layouts:           desc.Layouts,
			EndorsersByGroups: endorsersByGroups,
		})
	}
	return res
}

func endorserFromRaw(p *discovery.Peer) endorser {
	sId := &msp.SerializedIdentity{}
	proto.Unmarshal(p.Identity, sId)
	return endorser{
		MSPID:        sId.Mspid,
		Endpoint:     endpointFromEnvelope(p.MembershipInfo),
		LedgerHeight: ledgerHeightFromEnvelope(p.StateInfo),
		Identity:     string(sId.IdBytes),
	}
}

func endpointFromEnvelope(env *gossip.Envelope) string {
	if env == nil {
		return ""
	}
	aliveMsg, _ := env.ToGossipMessage()
	if aliveMsg == nil {
		return ""
	}
	if !aliveMsg.IsAliveMsg() {
		return ""
	}
	if aliveMsg.GetAliveMsg().Membership == nil {
		return ""
	}
	return aliveMsg.GetAliveMsg().Membership.Endpoint
}

func ledgerHeightFromEnvelope(env *gossip.Envelope) uint64 {
	if env == nil {
		return 0
	}
	stateInfoMsg, _ := env.ToGossipMessage()
	if stateInfoMsg == nil {
		return 0
	}
	if !stateInfoMsg.IsStateInfoMsg() {
		return 0
	}
	if stateInfoMsg.GetStateInfo().Properties == nil {
		return 0
	}
	return stateInfoMsg.GetStateInfo().Properties.LedgerHeight
}
