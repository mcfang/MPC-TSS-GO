package demo

import (
	"fmt"
	"path/filepath"
	"runtime"
)

const (
	testFixtureDirFormat  = "%s/_fixtures"
	testFixtureFileFormat = "keygen_data_%d.json"
	qty                   = 3
	testThreshold         = 2
)

func makeTestFixtureFilePath(partyIndex int) string {
	_, callerFileName, _, _ := runtime.Caller(0)
	srcDirName := filepath.Dir(callerFileName)
	fixtureDirName := fmt.Sprintf(testFixtureDirFormat, srcDirName)
	return fmt.Sprintf("%s/"+testFixtureFileFormat, fixtureDirName, partyIndex)
}
