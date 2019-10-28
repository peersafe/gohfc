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

func NewTLSCertInfo(cryptoSuite CryptoSuite, certFile, keyFile string) error {
	if nil == cryptoSuite {
		return errors.New("the cryptoSuite is empty")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	baseTLSCertInfo.certHash = cryptoSuite.Hash(cert.Certificate[0])
	baseTLSCertInfo.cert = cert

	return nil
}
