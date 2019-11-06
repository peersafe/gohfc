/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
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
func (p *Peer) Endorse(resp chan *PeerResponse, prop *peer.SignedProposal) error {
	proposalResp, err := p.client.ProcessProposal(context.Background(), prop)
	if err != nil {
		resp <- &PeerResponse{Response: nil, Name: p.Name, Err: err}
		return err
	}
	resp <- &PeerResponse{Response: proposalResp, Name: p.Name, Err: nil}
	return nil
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

func NewConnection(conf *ConnectionConfig) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if !conf.UseTLS {
		opts = append(opts, grpc.WithInsecure())
	}

	if conf.TlsMutual {
		//cert, err := tls.LoadX509KeyPair(conf.ClientCert, conf.ClientKey)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to Load client keypair: %s\n", err.Error())
		//}
		caPem, err := ioutil.ReadFile(conf.TlsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert file %s\n", conf.TlsPath)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caPem)
		c := &tls.Config{
			ServerName:   conf.DomainName,
			MinVersion:   tls.VersionTLS12, // GO1.12 adds opt-in support for TLS 1.3, and fabric 1.4 uses TLS 1.2
			Certificates: []tls.Certificate{baseTLSCertInfo.cert},
			RootCAs:      certPool,
			//InsecureSkipVerify: true, // Client verifies server's cert if false, else skip.
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c)))
	} else {
		var creds credentials.TransportCredentials
		var err error
		if conf.TlsPath != "" {
			creds, err = credentials.NewClientTLSFromFile(conf.TlsPath, conf.DomainName)
			if err != nil {
				return nil, fmt.Errorf("cannot read credentials %s and err is: %v", conf.TlsPath, err)
			}
		} else {
			certPool := x509.NewCertPool()
			if isCorrect := certPool.AppendCertsFromPEM(conf.TLSInfo[0]); !isCorrect {
				logger.Errorf("append certs failed for %s", conf.Host)
				return nil, errors.New("append certs failed")
			}
			creds = credentials.NewClientTLSFromCert(certPool, "")
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(1) * time.Minute,
			Timeout:             time.Duration(20) * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize)),
		grpc.WithTimeout(time.Second*30),
		grpc.WithBackoffConfig(grpc.BackoffConfig{
			MaxDelay: time.Second * 10,
		}))

	conn, err := grpc.Dial(conf.Host, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewPeerFromConn(c *grpc.ClientConn) peer.EndorserClient {
	return peer.NewEndorserClient(c)
}
