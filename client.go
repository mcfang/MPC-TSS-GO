package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"sync/atomic"

	"github.com/bnb-chain/tss-lib/common"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/test"
	"github.com/bnb-chain/tss-lib/tss"
)

func clientInit() (resPartyIDs tss.UnSortedPartyIDs, resKeys []keygen.LocalPartySaveData) {
	keys := make([]keygen.LocalPartySaveData, 0, qty)
	start := 0
	for i := start; i < qty-1; i++ {
		fixtureFilePath := makeTestFixtureFilePath(i)
		bz, err := ioutil.ReadFile(fixtureFilePath)
		if err != nil {
			return
		}
		var key keygen.LocalPartySaveData
		if err = json.Unmarshal(bz, &key); err != nil {
			return
		}
		for _, kbxj := range key.BigXj {
			kbxj.SetCurve(tss.S256())
		}
		key.ECDSAPub.SetCurve(tss.S256())
		keys = append(keys, key)
	}
	partyIDs := make(tss.UnSortedPartyIDs, len(keys))
	for i, key := range keys {
		pMoniker := fmt.Sprintf("%d", i+start+1)
		partyIDs[i] = tss.NewPartyID(pMoniker, pMoniker, key.ShareID)
	}
	return partyIDs, keys
}

func clientRun(partyIDs tss.UnSortedPartyIDs, keys []keygen.LocalPartySaveData) {
	pIDs := tss.SortPartyIDs(partyIDs)
	p2pCtx := tss.NewPeerContext(pIDs)
	parties := make([]*keygen.LocalParty, 0, len(pIDs))

	errCh := make(chan *tss.Error, len(pIDs))
	outCh := make(chan tss.Message, len(pIDs))
	endCh := make(chan keygen.LocalPartySaveData, len(pIDs))

	updater := test.SharedPartyUpdater

	startGR := runtime.NumGoroutine()

	// init the parties
	for i := 0; i < len(pIDs)-1; i++ {
		var P *keygen.LocalParty
		params := tss.NewParameters(tss.S256(), p2pCtx, pIDs[i], len(pIDs), testThreshold)
		P = keygen.NewLocalParty(params, outCh, endCh, keys[i].LocalPreParams).(*keygen.LocalParty)
		parties = append(parties, P)
		go func(P *keygen.LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	// PHASE: keygen
	var ended int32
keygen:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			common.Logger.Errorf("Error: %s", err)
			break keygen

		case msg := <-outCh:
			dest := msg.GetTo()
			if dest == nil { // broadcast!
				for _, P := range parties {
					if P.PartyID().Index == msg.GetFrom().Index {
						continue
					}
					go updater(P, msg, errCh)
				}
			} else { // point-to-point!
				if dest[0].Index == msg.GetFrom().Index {
					fmt.Errorf("party %d tried to send a message to itself (%d)", dest[0].Index, msg.GetFrom().Index)
					return
				}
				go updater(parties[dest[0].Index], msg, errCh)
			}

		case save := <-endCh:
			// SAVE a test fixture file for this P (if it doesn't already exist)
			// .. here comes a workaround to recover this party's index (it was removed from save data)
			index, _ := save.OriginalIndex()
			fmt.Printf("Party %d private key share (Xi): %s\n", index, save.LocalSecrets.Xi.String())
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(pIDs)) {
				fmt.Printf("Done. Received save data from %d participants", ended)
				fmt.Printf("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break keygen
			}
		}
	}
}
