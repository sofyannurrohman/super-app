package usecase

import (
	"sync"
	"time"
)

var (
	createTxnMu sync.Mutex
	activeTxns  = make(map[string]time.Time)
)

func lockReceiptNo(receiptNo string) bool {
	if receiptNo == "" {
		return true // skip lock if empty
	}
	createTxnMu.Lock()
	defer createTxnMu.Unlock()

	// cleanup old
	now := time.Now()
	for k, v := range activeTxns {
		if now.Sub(v) > 2*time.Minute {
			delete(activeTxns, k)
		}
	}

	if _, exists := activeTxns[receiptNo]; exists {
		return false // duplicate in progress
	}
	activeTxns[receiptNo] = now
	return true
}

func unlockReceiptNo(receiptNo string) {
	if receiptNo == "" {
		return
	}
	createTxnMu.Lock()
	defer createTxnMu.Unlock()
	delete(activeTxns, receiptNo)
}
