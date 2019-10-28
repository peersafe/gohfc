package parseBlock

import (
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	ledgerUtil "github.com/hyperledger/fabric/core/ledger/util"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	pbmsp "github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var parseBlockLogger = logging.MustGetLogger("sdk")

func deserializeIdentity(serializedID []byte) (*x509.Certificate, error) {
	sId := &pbmsp.SerializedIdentity{}
	err := proto.Unmarshal(serializedID, sId)
	if err != nil {
		return nil, fmt.Errorf("Could not deserialize a SerializedIdentity, err %s", err)
	}

	bl, _ := pem.Decode(sId.IdBytes)
	if bl == nil {
		return nil, fmt.Errorf("Could not decode the PEM structure")
	}
	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ParseCertificate failed %s", err)
	}
	return cert, nil
}

func copyChannelHeaderToLocalChannelHeader(localChannelHeader *ChannelHeader,
	chHeader *cb.ChannelHeader, chaincodeHeaderExtension *peer.ChaincodeHeaderExtension) {
	localChannelHeader.Type = chHeader.Type
	localChannelHeader.Version = chHeader.Version
	localChannelHeader.Timestamp = chHeader.Timestamp
	localChannelHeader.ChannelId = chHeader.ChannelId
	localChannelHeader.TxId = chHeader.TxId
	localChannelHeader.Epoch = chHeader.Epoch
	localChannelHeader.ChaincodeId = chaincodeHeaderExtension.ChaincodeId
}

func copyChaincodeSpecToLocalChaincodeSpec(localChaincodeSpec *ChaincodeSpec, chaincodeSpec *peer.ChaincodeSpec) {
	localChaincodeSpec.Type = chaincodeSpec.Type
	localChaincodeSpec.ChaincodeId = chaincodeSpec.ChaincodeId
	localChaincodeSpec.Timeout = chaincodeSpec.Timeout
	chaincodeInput := &ChaincodeInput{}
	for _, input := range chaincodeSpec.Input.Args {
		chaincodeInput.Args = append(chaincodeInput.Args, string(input))
	}
	localChaincodeSpec.Input = chaincodeInput
}

func copyEndorsementToLocalEndorsement(localTransaction *Transaction, allEndorsements []*peer.Endorsement) {
	for _, endorser := range allEndorsements {
		endorsement := &Endorsement{}
		endorserSignatureHeader := &cb.SignatureHeader{}

		endorserSignatureHeader.Creator = endorser.Endorser
		endorsement.SignatureHeader = getSignatureHeaderFromBlockData(endorserSignatureHeader)
		endorsement.Signature = endorser.Signature
		localTransaction.Endorsements = append(localTransaction.Endorsements, endorsement)
	}
}

func getValueFromBlockMetadata(block *cb.Block, index cb.BlockMetadataIndex) []byte {
	valueMetadata := &cb.Metadata{}
	if index == cb.BlockMetadataIndex_LAST_CONFIG {
		if err := proto.Unmarshal(block.Metadata.Metadata[index], valueMetadata); err != nil {
			return nil
		}

		lastConfig := &cb.LastConfig{}
		if err := proto.Unmarshal(valueMetadata.Value, lastConfig); err != nil {
			return nil
		}
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(lastConfig.Index))
		return b
	} else if index == cb.BlockMetadataIndex_ORDERER {
		if err := proto.Unmarshal(block.Metadata.Metadata[index], valueMetadata); err != nil {
			return nil
		}

		kafkaMetadata := &ab.KafkaMetadata{}
		if err := proto.Unmarshal(valueMetadata.Value, kafkaMetadata); err != nil {
			return nil
		}
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(kafkaMetadata.LastOffsetPersisted))
		return b
	} else if index == cb.BlockMetadataIndex_TRANSACTIONS_FILTER {
		return block.Metadata.Metadata[index]
	}
	return valueMetadata.Value
}

func getSignatureHeaderFromBlockMetadata(block *cb.Block, index cb.BlockMetadataIndex) (*SignatureMetadata, error) {
	signatureMetadata := &cb.Metadata{}
	if err := proto.Unmarshal(block.Metadata.Metadata[index], signatureMetadata); err != nil {
		return nil, err
	}
	localSignatureHeader := &cb.SignatureHeader{}

	if len(signatureMetadata.Signatures) > 0 {
		if err := proto.Unmarshal(signatureMetadata.Signatures[0].SignatureHeader, localSignatureHeader); err != nil {
			return nil, err
		}

		localSignatureMetadata := &SignatureMetadata{}
		localSignatureMetadata.SignatureHeader = getSignatureHeaderFromBlockData(localSignatureHeader)
		localSignatureMetadata.Signature = signatureMetadata.Signatures[0].Signature

		return localSignatureMetadata, nil
	}
	return nil, nil
}

func getSignatureHeaderFromBlockData(header *cb.SignatureHeader) *SignatureHeader {
	signatureHeader := &SignatureHeader{}
	signatureHeader.Certificate, _ = deserializeIdentity(header.Creator)
	signatureHeader.Nonce = header.Nonce
	return signatureHeader

}

// This method add transaction validation information from block TransactionFilter struct
func addTransactionValidation(block *Block, tran *Transaction, txIdx int) error {
	if len(block.TransactionFilter) > txIdx {
		tran.ValidationCode = uint8(block.TransactionFilter[txIdx])
		tran.ValidationCodeName = peer.TxValidationCode_name[int32(tran.ValidationCode)]
		return nil
	}
	return fmt.Errorf("invalid index or transaction filler. Index: %d", txIdx)
}

type BlockPerf struct {
	BlockNumber       int
	NumValidTx        int
	NumInvalidTx      int
	BlockDurationNs   int64 // Difference between block receving time and the time of submission of first proposal in block
	TxValidationStats map[string]int
	TxPerfs           []TxPerf
}

type TxPerf struct {
	TxId                   string
	ProposalSubmissionTime time.Time
	TxCommitTime           time.Time
	LatencyNs              int64 // latency in nanoseconds
}

func processBlock(block *cb.Block, size uint64) Block {
	var localBlock Block

	localBlock.Size = size
	localBlock.BlockTimeStamp = time.Now().UTC()

	localBlock.Header = block.Header
	localBlock.TransactionFilter = ledgerUtil.NewTxValidationFlags(len(block.Data.Data))

	// process block metadata before data
	localBlock.BlockCreatorSignature, _ = getSignatureHeaderFromBlockMetadata(block, cb.BlockMetadataIndex_SIGNATURES)
	lastConfigBlockNumber := &LastConfigMetadata{}
	lastConfigBlockNumber.LastConfigBlockNum = binary.LittleEndian.Uint64(getValueFromBlockMetadata(block, cb.BlockMetadataIndex_LAST_CONFIG))
	lastConfigBlockNumber.SignatureData, _ = getSignatureHeaderFromBlockMetadata(block, cb.BlockMetadataIndex_LAST_CONFIG)
	localBlock.LastConfigBlockNumber = lastConfigBlockNumber

	txBytes := getValueFromBlockMetadata(block, cb.BlockMetadataIndex_TRANSACTIONS_FILTER)
	for index, b := range txBytes {
		localBlock.TransactionFilter[index] = b
	}

	ordererKafkaMetadata := &OrdererMetadata{}
	ordererKafkaMetadata.LastOffsetPersisted = binary.BigEndian.Uint64(getValueFromBlockMetadata(block, cb.BlockMetadataIndex_ORDERER))
	ordererKafkaMetadata.SignatureData, _ = getSignatureHeaderFromBlockMetadata(block, cb.BlockMetadataIndex_ORDERER)
	localBlock.OrdererKafkaMetadata = ordererKafkaMetadata

	for txIndex, data := range block.Data.Data {
		localTransaction := &Transaction{}
		localTransaction.Size = uint64(len(data))
		//Get envelope which is stored as byte array in the data field.
		envelope, err := utils.GetEnvelopeFromBlock(data)
		if err != nil {
			parseBlockLogger.Errorf("Error getting envelope: %s\n", err)
			continue
		}
		localTransaction.Signature = envelope.Signature
		//Get payload from envelope struct which is stored as byte array.
		payload, err := utils.GetPayload(envelope)
		if err != nil {
			parseBlockLogger.Errorf("Error getting payload from envelope: %s\n", err)
			continue
		}
		chHeader, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			parseBlockLogger.Errorf("Error unmarshaling channel header: %s\n", err)
			continue
		}
		headerExtension := &peer.ChaincodeHeaderExtension{}
		if err := proto.Unmarshal(chHeader.Extension, headerExtension); err != nil {
			parseBlockLogger.Errorf("Error unmarshaling chaincode header extension: %s\n", err)
			continue
		}
		localChannelHeader := &ChannelHeader{}
		copyChannelHeaderToLocalChannelHeader(localChannelHeader, chHeader, headerExtension)

		localBlock.ChannelID = localChannelHeader.ChannelId
		if txIndex == 0 {
			localBlock.FirstTxTime = time.Unix(localChannelHeader.Timestamp.Seconds, int64(localChannelHeader.Timestamp.Nanos)).UTC()
		}

		// Performance measurement code ends
		localTransaction.ChannelHeader = localChannelHeader
		localSignatureHeader := &cb.SignatureHeader{}
		if err := proto.Unmarshal(payload.Header.SignatureHeader, localSignatureHeader); err != nil {
			parseBlockLogger.Errorf("Error unmarshaling signature header: %s\n", err)
			continue
		}
		localTransaction.SignatureHeader = getSignatureHeaderFromBlockData(localSignatureHeader)
		//localTransaction.SignatureHeader.Nonce = localSignatureHeader.Nonce
		//localTransaction.SignatureHeader.Certificate, _ = deserializeIdentity(localSignatureHeader.Creator)

		if cb.HeaderType(chHeader.Type) == cb.HeaderType_ENDORSER_TRANSACTION {
			transaction := &peer.Transaction{}
			if err := proto.Unmarshal(payload.Data, transaction); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling transaction: %s\n", err)
				continue
			}
			chaincodeActionPayload, chaincodeAction, err := utils.GetPayloads(transaction.Actions[0])
			if err != nil {
				parseBlockLogger.Errorf("Error getting payloads from transaction actions: %s\n", err)
				continue
			}
			localSignatureHeader = &cb.SignatureHeader{}
			if err := proto.Unmarshal(transaction.Actions[0].Header, localSignatureHeader); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling signature header: %s\n", err)
				continue
			}
			localTransaction.TxActionSignatureHeader = getSignatureHeaderFromBlockData(localSignatureHeader)
			//signatureHeader = &SignatureHeader{}
			//signatureHeader.Certificate, _ = deserializeIdentity(localSignatureHeader.Creator)
			//signatureHeader.Nonce = localSignatureHeader.Nonce
			//localTransaction.TxActionSignatureHeader = signatureHeader

			chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
			if err := proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling chaincode proposal payload: %s\n", err)
				continue
			}
			chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{}
			if err := proto.Unmarshal(chaincodeProposalPayload.Input, chaincodeInvocationSpec); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling chaincode invocationSpec: %s\n", err)
				continue
			}
			localChaincodeSpec := &ChaincodeSpec{}
			copyChaincodeSpecToLocalChaincodeSpec(localChaincodeSpec, chaincodeInvocationSpec.ChaincodeSpec)
			localTransaction.ChaincodeSpec = localChaincodeSpec
			copyEndorsementToLocalEndorsement(localTransaction, chaincodeActionPayload.Action.Endorsements)
			proposalResponsePayload := &peer.ProposalResponsePayload{}
			if err := proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling proposal response payload: %s\n", err)
				continue
			}
			localTransaction.ProposalHash = proposalResponsePayload.ProposalHash
			localTransaction.Response = chaincodeAction.Response
			events := &peer.ChaincodeEvent{}
			if err := proto.Unmarshal(chaincodeAction.Events, events); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling chaincode action events:%s\n", err)
				continue
			}
			localTransaction.Events = events

			txReadWriteSet := &rwset.TxReadWriteSet{}
			if err := proto.Unmarshal(chaincodeAction.Results, txReadWriteSet); err != nil {
				parseBlockLogger.Errorf("Error unmarshaling chaincode action results: %s\n", err)
				continue
			}

			if len(chaincodeAction.Results) != 0 {
				for _, nsRwset := range txReadWriteSet.NsRwset {
					nsReadWriteSet := &NsReadWriteSet{}
					kvRWSet := &kvrwset.KVRWSet{}
					nsReadWriteSet.Namespace = nsRwset.Namespace
					if err := proto.Unmarshal(nsRwset.Rwset, kvRWSet); err != nil {
						parseBlockLogger.Errorf("Error unmarshaling tx read write set: %s\n", err)
						continue
					}
					nsReadWriteSet.KVRWSet = kvRWSet
					localTransaction.NsRwset = append(localTransaction.NsRwset, nsReadWriteSet)
				}
			}

			// add the transaction validation a
			addTransactionValidation(&localBlock, localTransaction, txIndex)
			//append the transaction
			localBlock.Transactions = append(localBlock.Transactions, localTransaction)
		} else if cb.HeaderType(chHeader.Type) == cb.HeaderType_CONFIG {
			configEnv := &cb.ConfigEnvelope{}
			_, err = utils.UnmarshalEnvelopeOfType(envelope, cb.HeaderType_CONFIG, configEnv)
			if err != nil {
				parseBlockLogger.Errorf("Bad configuration envelope: %s", err)
				continue
			}

			parseBlockLogger.Debugf("it's config block number : %d", block.Header.Number)
		}

	}

	return localBlock
}

func ParseBlock(blockEvent *cb.Block, size uint64) Block {
	return processBlock(blockEvent, size)
}
