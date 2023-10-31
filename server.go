package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
)

func serverInit() (resPartyIDs tss.UnSortedPartyIDs, resKeys []keygen.LocalPartySaveData) {
	keys := make([]keygen.LocalPartySaveData, 0, qty)
	start := 2
	for i := start; i < qty; i++ {
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
