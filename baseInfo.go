package gohfc

import (
	"crypto/tls"
	"errors"
)

type tlsCertInfo struct {
	cert     tls.Certificate
	certHash []byte
}

var baseTLSCertInfo tlsCertInfo

func NewTLSCertInfo(fabClient *FabricClient) error {
	if nil == fabClient.Crypto {
		return errors.New("the cryptoSuite is empty")
	}
	cert, err := tls.LoadX509KeyPair(fabClient.ClientCert, fabClient.ClientKey)
	if err != nil {
		return err
	}

	baseTLSCertInfo.certHash = fabClient.Crypto.Hash(cert.Certificate[0])
	baseTLSCertInfo.cert = cert

	return nil
}
