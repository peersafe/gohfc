/*
Copyright: Cognition Foundry. All Rights Reserved.
License: Apache License Version 2.0
*/
package gohfc

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ClientConfig holds config data for crypto, peers and orderers
type ClientConfig struct {
	CryptoConfig   `yaml:"crypto"`
	Orderers       map[string]OrdererConfig    `yaml:"orderers"`
	Peers          map[string]PeerConfig       `yaml:"peers"`
	EventPeers     map[string]PeerConfig       `yaml:"eventPeers"`
	DiscoveryPeers map[string]ConnectionConfig `yaml:"discoveryPeers"`
	ChannelConfig  ChannelConfig  `yaml:"channel"`
}

type ChannelConfig struct {
	MspConfigPath    string `yaml:"mspConfigPath"`
	LocalMspId       string `yaml:"localMspId"`
	ChannelId        string `yaml:"channelId"`
	ChaincodeName    []string `yaml:"chaincodeName"`
//	ChaincodeVersion string `yaml:"chaincodeVersion"`
//	ChaincodePolicy  `yaml:"chaincodePolicy"`
	ClientCert       string `yaml:"clientCert"`
	ClientKey        string `yaml:"clientKey"`
}

type ChaincodePolicy struct {
	Orgs []string `yaml:"orgs"`
	Rule string   `yaml:"rule"`
}

// CAConfig holds config for Fabric CA
type CAConfig struct {
	CryptoConfig      `yaml:"crypto"`
	Uri               string `yaml:"url"`
	SkipTLSValidation bool   `yaml:"skipTLSValidation"`
	MspId             string `yaml:"mspId"`
}

// Config holds config values for fabric and fabric-ca cryptography
type CryptoConfig struct {
	Family    string `yaml:"family"`
	Algorithm string `yaml:"algorithm"`
	Hash      string `yaml:"hash"`
}

// PeerConfig hold config values for Peer. ULR is in address:port notation
type PeerConfig struct {
	Host       string `yaml:"host"`
	OrgName    string `yaml:"orgName"`
	UseTLS     bool   `yaml:"useTLS"`
	TlsPath    string `yaml:"tlsPath"`
	TLSInfo    [][]byte
	DomainName string `yaml:"domainName"`
	TlsMutual  bool   `yaml:"tlsMutual"`
	ClientCert string `yaml:"clientCert"`
	ClientKey  string `yaml:"clientKey"`
}

// OrdererConfig hold config values for Orderer. ULR is in address:port notation
type OrdererConfig struct {
	Host       string `yaml:"host"`
	UseTLS     bool   `yaml:"useTLS"`
	TlsPath    string `yaml:"tlsPath"`
	TLSInfo    [][]byte
	DomainName string `yaml:"domainName"`
	TlsMutual  bool   `yaml:"tlsMutual"`
	ClientCert string `yaml:"clientCert"`
	ClientKey  string `yaml:"clientKey"`
}

type ConnectionConfig struct {
	Host       string `yaml:"host"`
	UseTLS     bool   `yaml:"useTLS"`
	TlsPath    string `yaml:"tlsPath"`
	TLSInfo    [][]byte
	DomainName string `yaml:"domainName"`
	TlsMutual  bool   `yaml:"tlsMutual"`
	ClientCert string `yaml:"clientCert"`
	ClientKey  string `yaml:"clientKey"`
}

// NewFabricClientConfig create config from provided yaml file in path
func NewClientConfig(path string) (*ClientConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(ClientConfig)
	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// NewCAConfig create new Fabric CA config from provided yaml file in path
func NewCAConfig(path string) (*CAConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(CAConfig)
	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
