package waitTxstatus

import (
	"fmt"
	"sync"
)

//wait transaction status

var (
	GlobalTxStatusMap sync.Map
	GlobalChan        = make(chan string, 10)
)

func RegisterTxStatusEvent(channelID, txID string) (chan string, error) {
	if txID == "" {
		return nil, fmt.Errorf("txID must be provided")
	}
	var newChanTxStatus sync.Map
	factChanTxStatusMap, ok := GlobalTxStatusMap.LoadOrStore(channelID, &newChanTxStatus)
	if !ok {
		GlobalChan <- channelID
	}
	statusChan := make(chan string)
	if v, ok := factChanTxStatusMap.(*sync.Map); ok {
		_, exist := v.LoadOrStore(txID, statusChan)
		if exist {
			return nil, fmt.Errorf("TxStatusEvent txID %s was exist", txID)
		}
	} else {
		return nil, fmt.Errorf("factChanTxStatusMap convert failed")
	}
	return statusChan, nil
}

func UnRegisterTxStatusEvent(channelID, txID string, statusChan chan string) {
	close(statusChan)
	if v, exist := GlobalTxStatusMap.Load(channelID); exist {
		if txStatusMap, ok := v.(*sync.Map); ok {
			txStatusMap.Delete(txID)
		}
	}
}

func PublishTxStatus(channelID, txID string, txStatus string) {
	go func(chanName, id string, status string) {
		if v, exist := GlobalTxStatusMap.Load(chanName); exist {
			if txStatusMap, ok := v.(*sync.Map); ok {
				statusChan, exist := txStatusMap.Load(id)
				if exist {
					if ch, ok := statusChan.(chan string); ok {
						ch <- status
					}
				}
			}
		}
	}(channelID, txID, txStatus)
}
