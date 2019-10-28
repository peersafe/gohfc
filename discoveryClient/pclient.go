package discoveryClient

import (
	"context"

	"github.com/hyperledger/fabric/protos/discovery"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type PClient struct {
	Name        string
	uri         string
	client      discovery.DiscoveryClient
	signRequest Signer
}

// NewClient creates a new Client instance
func NewPClient(uri, name string, cn *grpc.ClientConn, s Signer, signerCacheSize uint) *PClient {
	return &PClient{
		uri:         uri,
		Name:        name,
		client:      discovery.NewDiscoveryClient(cn),
		signRequest: NewMemoizeSigner(s, signerCacheSize).Sign,
	}
}

func (pc *PClient) Send(ctx context.Context, req *Request, auth *discovery.AuthInfo) (Response, error) {
	reqToBeSent := *req.Request
	reqToBeSent.Authentication = auth
	payload, err := proto.Marshal(&reqToBeSent)
	if err != nil {
		return nil, errors.Wrap(err, "failed marshaling Request to bytes")
	}

	sig, err := pc.signRequest(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed sign payload")
	}

	resp, err := pc.client.Discover(ctx, &discovery.SignedRequest{
		Payload:   payload,
		Signature: sig,
	})
	if err != nil {
		return nil, errors.Wrap(err, "discovery service refused our Request")
	}

	if n := len(resp.Results); n != req.lastIndex {
		return nil, errors.Errorf("Sent %d queries but received %d responses back", req.lastIndex, n)
	}
	return req.computeResponse(resp)
}

func (pc *PClient) RowSend(ctx context.Context, req *Request, auth *discovery.AuthInfo) (*discovery.Response, error) {
	reqToBeSent := *req.Request
	reqToBeSent.Authentication = auth
	payload, err := proto.Marshal(&reqToBeSent)
	if err != nil {
		return nil, errors.Wrap(err, "failed marshaling Request to bytes")
	}

	sig, err := pc.signRequest(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed sign payload")
	}

	resp, err := pc.client.Discover(ctx, &discovery.SignedRequest{
		Payload:   payload,
		Signature: sig,
	})
	if err != nil {
		return nil, errors.Wrap(err, "discovery service refused our Request")
	}

	return resp, nil
}
