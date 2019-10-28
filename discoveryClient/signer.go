/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryClient

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protos/msp"
	proto_utils "github.com/hyperledger/fabric/protos/utils"
)

// Signer signs messages.
// TODO: Ideally we'd use an MSP to be agnostic, but since it's impossible to
// initialize an MSP without a CA cert that signs the signing identity,
// this will do for now.
type Sign struct {
	key     *ecdsa.PrivateKey
	Creator []byte
}

func NewSign(mspPath, mspID string) *Sign {
	if mspPath == "" || mspID == "" {
		return nil
	}
	cert, priKey, err := findCertAndKeyFile(mspPath)
	if err != nil {
		return nil
	}

	sId, err := serializeIdentity(cert, mspID)
	if err != nil {
		return nil
	}
	key, err := loadPrivateKey(priKey)
	if err != nil {
		return nil
	}

	return &Sign{
		Creator: sId,
		key:     key,
	}
}

func serializeIdentity(clientCert string, mspID string) ([]byte, error) {
	b, err := ioutil.ReadFile(clientCert)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sId := &msp.SerializedIdentity{
		Mspid:   mspID,
		IdBytes: b,
	}
	return proto_utils.MarshalOrPanic(sId), nil
}

func loadPrivateKey(file string) (*ecdsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	bl, _ := pem.Decode(b)
	if bl == nil {
		return nil, errors.Errorf("failed to decode PEM block from %s", file)
	}
	key, err := x509.ParsePKCS8PrivateKey(bl.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse private key from %s", file)
	}
	return key.(*ecdsa.PrivateKey), nil
}

func findCertAndKeyFile(msppath string) (string, string, error) {
	findCert := func(path string) (string, error) {
		list, err := ioutil.ReadDir(path)
		if err != nil {
			return "", err
		}
		var file os.FileInfo
		for _, item := range list {
			if !item.IsDir() {
				if file == nil {
					file = item
				} else if item.ModTime().After(file.ModTime()) {
					file = item
				}
			}
		}
		if file == nil {
			return "", fmt.Errorf("have't file in the %s", path)
		}
		return filepath.Join(path, file.Name()), nil
	}
	prikey, err := findCert(filepath.Join(msppath, "keystore"))
	if err != nil {
		return "", "", err
	}
	cert, err := findCert(filepath.Join(msppath, "signcerts"))
	if err != nil {
		return "", "", err
	}
	return cert, prikey, nil
}

func (si *Sign) Sign(msg []byte) ([]byte, error) {
	digest := util.ComputeSHA256(msg)
	return signECDSA(si.key, digest)
}

func signECDSA(k *ecdsa.PrivateKey, digest []byte) (signature []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, k, digest)
	if err != nil {
		return nil, err
	}

	s, _, err = utils.ToLowS(&k.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return marshalECDSASignature(r, s)
}

func marshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

type ECDSASignature struct {
	R, S *big.Int
}

// MemoizeSigner signs messages with the same signature
// if the message was signed recently
type MemoizeSigner struct {
	maxEntries uint
	sync.RWMutex
	memory map[string][]byte
	sign   Signer
}

// NewMemoizeSigner creates a new MemoizeSigner that signs
// message with the given sign function
func NewMemoizeSigner(signFunc Signer, maxEntries uint) *MemoizeSigner {
	return &MemoizeSigner{
		maxEntries: maxEntries,
		memory:     make(map[string][]byte),
		sign:       signFunc,
	}
}

// Signer signs a message and returns the signature and nil,
// or nil and error on failure
func (ms *MemoizeSigner) Sign(msg []byte) ([]byte, error) {
	sig, isInMemory := ms.lookup(msg)
	if isInMemory {
		return sig, nil
	}
	sig, err := ms.sign(msg)
	if err != nil {
		return nil, err
	}
	ms.memorize(msg, sig)
	return sig, nil
}

// lookup looks up the given message in memory and returns
// the signature, if the message is in memory
func (ms *MemoizeSigner) lookup(msg []byte) ([]byte, bool) {
	ms.RLock()
	defer ms.RUnlock()
	sig, exists := ms.memory[msgDigest(msg)]
	return sig, exists
}

func (ms *MemoizeSigner) memorize(msg, signature []byte) {
	if ms.maxEntries == 0 {
		return
	}
	ms.RLock()
	shouldShrink := len(ms.memory) >= (int)(ms.maxEntries)
	ms.RUnlock()

	if shouldShrink {
		ms.shrinkMemory()
	}
	ms.Lock()
	defer ms.Unlock()
	ms.memory[msgDigest(msg)] = signature

}

// evict evicts random messages from memory
// until its size is smaller than maxEntries
func (ms *MemoizeSigner) shrinkMemory() {
	ms.Lock()
	defer ms.Unlock()
	for len(ms.memory) > (int)(ms.maxEntries) {
		ms.evictFromMemory()
	}
}

// evictFromMemory evicts a random message from memory
func (ms *MemoizeSigner) evictFromMemory() {
	for dig := range ms.memory {
		delete(ms.memory, dig)
		return
	}
}

// msgDigest returns a digest of a given message
func msgDigest(msg []byte) string {
	return hex.EncodeToString(util.ComputeSHA256(msg))
}
