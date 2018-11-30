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
	Name        string
	OrgName     string
	Uri         string
	MspId       string
	Opts        []grpc.DialOption
	caPath      string
	tlsCertHash []byte
	conn        *grpc.ClientConn
	client      peer.EndorserClient
}

// PeerResponse is response from peer transaction request
type PeerResponse struct {
	Response *peer.ProposalResponse
	Err      error
	Name     string
}

// Endorse sends single transaction to single peer.
func (p *Peer) Endorse(resp chan *PeerResponse, prop *peer.SignedProposal) {
	proposalResp, err := p.client.ProcessProposal(context.Background(), prop)
	if err != nil {
		resp <- &PeerResponse{Response: nil, Name: p.Name, Err: err}
		return
	}
	resp <- &PeerResponse{Response: proposalResp, Name: p.Name, Err: nil}
}

// NewPeerFromConfig creates new peer from provided config
func NewPeerFromConfig(conf PeerConfig, cryptoSuite CryptoSuite) (*Peer, error) {
	p := Peer{Uri: conf.Host, caPath: conf.TlsPath}
	if !conf.UseTLS {
		p.Opts = []grpc.DialOption{grpc.WithInsecure()}
	} else if p.caPath != "" {
		if conf.TlsMutual {
			cert, err := tls.LoadX509KeyPair(conf.ClientCert, conf.ClientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to Load client keypair: %s\n", err.Error())
			}
			if cryptoSuite != nil {
				p.tlsCertHash = cryptoSuite.Hash(cert.Certificate[0])
			}
			caPem, err := ioutil.ReadFile(conf.TlsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert file %s\n", conf.TlsPath)
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
			p.Opts = append(p.Opts, grpc.WithTransportCredentials(credentials.NewTLS(c)))
		} else {
			creds, err := credentials.NewClientTLSFromFile(p.caPath, conf.DomainName)
			if err != nil {
				return nil, fmt.Errorf("cannot read peer %s credentials err is: %v", p.Name, err)
			}
			p.Opts = append(p.Opts, grpc.WithTransportCredentials(creds))
		}
	}

	p.Opts = append(p.Opts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(1) * time.Minute,
			Timeout:             time.Duration(20) * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize)))

	conn, err := grpc.Dial(p.Uri, p.Opts...)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	p.client = peer.NewEndorserClient(p.conn)

	return &p, nil
}
