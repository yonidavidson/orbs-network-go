package native

import (
	"fmt"
	"github.com/orbs-network/orbs-network-go/test/contracts"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

const counterContractStartFrom = 100

func TestCompileCodeHappyFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compilation of contracts in short mode")
	}

	code := string(contracts.SourceCodeForCounter(counterContractStartFrom))
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	contractInfo, err := compileAndLoadDeployedSourceCode(code, tmpDir)
	require.NoError(t, err, "compile and load should succeed")
	require.NotNil(t, contractInfo, "loaded object should not be nil")
	require.Equal(t, fmt.Sprintf("CounterFrom%d", counterContractStartFrom), contractInfo.Name, "loaded object should be valid")
}

func TestCompileCodeWithExistingArtifacts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compilation of contracts in short mode")
	}

	code := string(contracts.SourceCodeForCounter(counterContractStartFrom))
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	t.Log("Build fresh artifacts")

	sourceFilePath, err := writeSourceCodeToDisk("testPrefix", code, tmpDir)
	require.NoError(t, err, "write to disk should succeed")
	require.FileExists(t, sourceFilePath, "file should exist")

	soFilePath, err := buildSharedObject("testPrefix", sourceFilePath, tmpDir)
	require.NoError(t, err, "compilation should succeed")
	require.FileExists(t, soFilePath, "file should exist")

	t.Log("Simulate corrupted artifacts and rebuild")

	// simulate corrupt file that exists
	err = ioutil.WriteFile(sourceFilePath, []byte{0x01}, 0644)
	require.NoError(t, err)
	require.Equal(t, int64(1), getFileSize(sourceFilePath), "file size should match")

	// simulate corrupt file that exists
	err = ioutil.WriteFile(soFilePath, []byte{0x01}, 0644)
	require.NoError(t, err)
	require.Equal(t, int64(1), getFileSize(soFilePath), "file size should match")

	sourceFilePath, err = writeSourceCodeToDisk("prefix", code, tmpDir)
	require.NoError(t, err, "write to disk should succeed")
	require.FileExists(t, sourceFilePath, "file should exist")
	require.NotEqual(t, int64(1), getFileSize(sourceFilePath), "file size should not match")

	soFilePath, err = buildSharedObject("testPrefix", sourceFilePath, tmpDir)
	require.NoError(t, err, "compilation should succeed")
	require.FileExists(t, soFilePath, "file should exist")
	require.NotEqual(t, int64(1), getFileSize(soFilePath), "file size should not match")

	t.Log("Load artifact")

	contractInfo, err := loadSharedObject(soFilePath)
	require.NoError(t, err, "load should succeed")
	require.NotNil(t, contractInfo, "loaded object should not be nil")
	require.Equal(t, fmt.Sprintf("CounterFrom%d", counterContractStartFrom), contractInfo.Name, "loaded object should be valid")

	t.Log("Try to rebuild already loaded artifact")

	soFilePath, err = buildSharedObject("testPrefix", sourceFilePath, tmpDir)
	require.NoError(t, err, "compilation should succeed")
	require.FileExists(t, soFilePath, "file should exist")
	require.NotEqual(t, int64(1), getFileSize(soFilePath), "file size should not match")

	contractInfo, err = loadSharedObject(soFilePath)
	require.NoError(t, err, "load should succeed")
	require.NotNil(t, contractInfo, "loaded object should not be nil")
	require.Equal(t, fmt.Sprintf("CounterFrom%d", counterContractStartFrom), contractInfo.Name, "loaded object should be valid")
}

func createTempTestDir(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("/tmp", t.Name())
	if err != nil {
		panic("could not create temp dir for test")
	}
	return tmpDir
}

func getFileSize(filePath string) int64 {
	fi, err := os.Stat(filePath)
	if err != nil {
		panic("could not get file size")
	}
	return fi.Size()
}
