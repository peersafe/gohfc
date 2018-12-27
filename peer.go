/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"time"
)

// Peer expose API's to communicate with peer
type Peer struct {
	Host        string
	MspId       string
	Opts        []grpc.DialOption
	tlsCertHash []byte
	conn        *grpc.ClientConn
	client      peer.EndorserClient
}

// PeerResponse is response from peer transaction request
type PeerResponse struct {
	Response *peer.ProposalResponse
	Err      error
	Host     string
}

// Endorse sends single transaction to single peer.
func (p *Peer) Endorse(resp chan *PeerResponse, prop *peer.SignedProposal) {
	proposalResp, err := p.client.ProcessProposal(context.Background(), prop)
	if err != nil {
		resp <- &PeerResponse{Response: nil, Host: p.Host, Err: err}
		return
	}
	resp <- &PeerResponse{Response: proposalResp, Host: p.Host, Err: nil}
}

// NewPeerFromConfig creates new peer from provided config
func NewPeerFromConfig(conf NodeConfig, cryptoSuite CryptoSuite) (*Peer, error) {
	p := Peer{Host: conf.Host}
	var err error
	p.Opts, p.tlsCertHash, err = GetOptsByConfig(conf, cryptoSuite)
	if err != nil {
		return nil, fmt.Errorf("connect host=%s failed, err:%s\n", p.Host, err.Error())
	}

	conn, err := grpc.Dial(conf.Host, p.Opts...)
	if err != nil {
		return nil, fmt.Errorf("connect host=%s failed, err:%s\n", p.Host, err.Error())
	}
	p.conn = conn
	p.client = peer.NewEndorserClient(p.conn)

	return &p, nil
}

func GetOptsByConfig(conf NodeConfig, cryptoSuite CryptoSuite) ([]grpc.DialOption, []byte, error) {
	var Opts []grpc.DialOption
	var tlsCertHash []byte
	if !conf.UseTls {
		Opts = []grpc.DialOption{grpc.WithInsecure()}
	} else if conf.TlsCaPath != "" {
		if conf.TlsMutual {
			cert, err := tls.LoadX509KeyPair(conf.TlsClientCert, conf.TlsClientKey)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to Load FabCli keypair: %s\n", err.Error())
			}
			if cryptoSuite != nil {
				tlsCertHash = cryptoSuite.Hash(cert.Certificate[0])
			}
			caPem, err := ioutil.ReadFile(conf.TlsCaPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read CA cert faild err:%s\n", err.Error())
			}
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(caPem)
			c := &tls.Config{
				ServerName:   conf.DomainName,
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
				RootCAs:      certpool,
				//InsecureSkipVerify: true, // Client verifies server's cert if false, else skip.
			}
			Opts = append(Opts, grpc.WithTransportCredentials(credentials.NewTLS(c)))
		} else {
			creds, err := credentials.NewClientTLSFromFile(conf.TlsCaPath, conf.DomainName)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot read peer %s credentials err is: %v", conf.Host, err)
			}
			Opts = append(Opts, grpc.WithTransportCredentials(creds))
		}
	}

	Opts = append(Opts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(1) * time.Minute,
			Timeout:             time.Duration(20) * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize)))

	return Opts, tlsCertHash, nil
}
