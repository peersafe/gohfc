package waitTxstatus

import (
	"fmt"
)

//wait transaction status

var (
	//logger            = logging.MustGetLogger("sdk")
	GlobalEventChan   = make(chan *TxStatusReg, 100)
	GlobalTxStatusMap *SyncMap
)

func init() {
	GlobalTxStatusMap = NewSyncMap(make(map[string]TxChan))
}

type TxStatusReg struct {
	TxID    string
	Eventch chan *TxStatusEvent
	Errch   chan error
}

// TxStatusEvent contains the data for a transaction status event
type TxStatusEvent struct {
	// TxID is the ID of the transaction in which the event was set
	// TxValidationCode is the status code of the commit
	TxValidationCode string
}

func RegisterTxStatusEvent(txID string) (*TxStatusReg, error) {
	if txID == "" {
		return nil, fmt.Errorf("txID must be provided")
	}
	statusChan := make(chan *TxStatusEvent)
	errChan := make(chan error)
	reg := &TxStatusReg{
		TxID:    txID,
		Eventch: statusChan,
		Errch:   errChan,
	}

	GlobalEventChan <- reg
	return reg, nil
}

func UnRegisterTxStatusEvent(reg *TxStatusReg) {
	close(reg.Eventch)
	GlobalTxStatusMap.Delete(reg.TxID)
}

func HandleRegisterTxStatusEvent() {
	go func() {
		for {
			e, ok := <-GlobalEventChan
			if !ok {
				break
			}
			_, exist := GlobalTxStatusMap.Put(e.TxID, e.Eventch)
			if exist {
				e.Errch <- fmt.Errorf("HandleRegisterTxStatusEvent txID %s was exist", e.TxID)
			}
		}
	}()
}

func PublishTxStatus(txID, txStatus string) {
	if txID == "" {
		return
	}
	go func(id string, status string) {
		statusChan, exist := GlobalTxStatusMap.Get(id)
		if exist {
			statusChan <- &TxStatusEvent{TxValidationCode: txStatus}
		}
	}(txID, txStatus)
}
