package waitTxstatus

import (
	"fmt"
	"sync"
)

//wait transaction status

var (
	GlobalTxStatusMap sync.Map
)

func RegisterTxStatusEvent(txID string) (chan string, error) {
	if txID == "" {
		return nil, fmt.Errorf("txID must be provided")
	}
	statusChan := make(chan string)
	_, exist := GlobalTxStatusMap.LoadOrStore(txID, statusChan)
	if exist {
		return nil, fmt.Errorf("TxStatusEvent txID %s was exist", txID)
	}
	return statusChan, nil
}

func UnRegisterTxStatusEvent(txID string, statusChan chan string) {
	close(statusChan)
	GlobalTxStatusMap.Delete(txID)
}

func PublishTxStatus(txID string, txStatus string) {
	go func(id string, status string) {
		statusChan, exist := GlobalTxStatusMap.Load(id)
		if exist {
			if ch, ok := statusChan.(chan string); ok {
				ch <- status
			}
		}
	}(txID, txStatus)
}
