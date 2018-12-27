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
type YamlConfig struct {
	CryptoConfig `yaml:"crypto"`
	Orderers     []NodeConfig `yaml:"orderers"`
	Peers        []NodeConfig `yaml:"peers"`
	MspConfig    `yaml:"msp"`
	Mq           `yaml:"mq"`
	Log          map[string]string `yaml:"log"`
}

type MspConfig struct {
	LocalMspId    string `yaml:"localMspId"`
	MspConfigPath string `yaml:"mspConfigPath"`
	ChannelId     string `yaml:"channelId"`
	ChaincodeName string `yaml:"chaincodeName"`
	TlsClientCert string `yaml:"tlsClientCert"`
	TlsClientKey  string `yaml:"tlsClientKey"`
}

type Mq struct {
	MqAddress []string `yaml:"mqAddress"`
	QueueName string   `yaml:"queueName"`
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

type NodeConfig struct {
	Host          string `yaml:"host"`
	DomainName    string `yaml:"domainName"`
	UseTls        bool   `yaml:"useTls"`
	TlsCaPath     string `yaml:"tlsCaPath"`
	TlsMutual     bool   `yaml:"tlsMutual"`
	TlsClientCert string `yaml:"tlsClientCert"`
	TlsClientKey  string `yaml:"tlsClientKey"`
}

// NewFabricClientConfig create config from provided yaml file in path
func NewYamlConfig(path string) (*YamlConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(YamlConfig)
	if err := yaml.Unmarshal([]byte(data), config); err != nil {
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
