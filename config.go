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
	LocalConfig    `yaml:"localConfig"`
	Orderers       map[string]OrdererConfig    `yaml:"orderers"`
	Peers          map[string]PeerConfig       `yaml:"peers"`
	DiscoveryPeers map[string]ConnectionConfig `yaml:"discoveryPeers"`
	CCofChannels   map[string][]string         `yaml:"ccofchannels"` //key为channelID，value为chaincodes
}

// EventConfig holds config data for event peers
type EventConfig struct {
	CryptoConfig `yaml:"crypto"`
	LocalConfig  `yaml:"localConfig"`
	//Orderers       map[string]OrdererConfig    `yaml:"orderers"`
	//Peers          map[string]PeerConfig       `yaml:"peers"`
	EventPeers map[string]ConnectionConfig `yaml:"eventPeers"`
	//DiscoveryPeers map[string]ConnectionConfig `yaml:"discoveryPeers"`
	Channel string `yaml:"channel"` //key为channelID，value为chaincodes
}

type LocalConfig struct {
	IsReConnect        bool   `yaml:"isReConnect"`
	ReConnTimeInterval int    `yaml:"reConnTimeInterval"`
	MspConfigPath      string `yaml:"mspConfigPath"`
	LocalMspId         string `yaml:"localMspId"`
	ClientCert         string `yaml:"clientCert"`
	ClientKey          string `yaml:"clientKey"`
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
	ChannelId  string
	MSPId      string
	TLSInfo    [][]byte
	DomainName string `yaml:"domainName"`
	TlsMutual  bool   `yaml:"tlsMutual"`
	ClientCert string `yaml:"clientCert"`
	ClientKey  string `yaml:"clientKey"`
}

// newClientConfig create config from provided yaml file in path
func newClientConfig(path string) (*ClientConfig, error) {
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

// newEventConfig create config from provided yaml file in path
func newEventConfig(path string) (*EventConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(EventConfig)
	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// NewCAConfig create new Fabric CA config from provided yaml file in path
func newCAConfig(path string) (*CAConfig, error) {
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
