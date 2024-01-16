// Copyright 2024 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package read_cache

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/client"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/log_parser/json_parser/read_logs"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/setup"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/test_setup"
	"log"
	"testing"
	"time"
)

const (
	MiB             = 1024 * 1024
	chunkSizeToRead = MiB
	fileSize        = 3 * MiB
	chunksRead      = fileSize / MiB
	testFileName    = "foo"
)

////////////////////////////////////////////////////////////////////////
// Boilerplate
////////////////////////////////////////////////////////////////////////

type testStruct struct {
	flags         []string
	storageClient *storage.Client
	ctx           context.Context
}

func (s *testStruct) Setup(t *testing.T) {
	mountBucket(s.flags,t)

	setup.SetMntDir(mountDir)
	testDirPath = client.SetupTestDirectory(s.ctx, s.storageClient, testDirName)
	client.SetupFileInTestDirectory(s.ctx, s.storageClient, testDirName, testFileName, fileSize, t)
}

func (s *testStruct) Teardown(t *testing.T) {
	unmountAndDeleteLogFile()
}

////////////////////////////////////////////////////////////////////////
// Test scenarios
////////////////////////////////////////////////////////////////////////

func (s *testStruct) TestSecondSequentialReadIsCacheHit(t *testing.T) {
	// Read file 1st time.
	expectedOutcome1 := readFileAndGetExpectedOutcome(testDirPath, testFileName, t)
	validateFileInCacheDirectory(s.ctx,s.storageClient,t)
	// Read file 2nd time.
	expectedOutcome2 := readFileAndGetExpectedOutcome(testDirPath, testFileName, t)

	// Validate that the content read by read operation matches content on GCS.
	client.ValidateObjectContentsFromGCS(s.ctx, s.storageClient, testDirName, testFileName,
		expectedOutcome1.content, t)
	client.ValidateObjectContentsFromGCS(s.ctx, s.storageClient, testDirName, testFileName,
		expectedOutcome2.content, t)
	// Parse the log file and validate cache hit or miss from the structured logs.
	structuredReadLogs := read_logs.GetStructuredLogsSortedByTimestamp(setup.LogFile(), t)
	validate(expectedOutcome1, structuredReadLogs[0], true, false, chunksRead, t)
	validate(expectedOutcome2, structuredReadLogs[1], true, true, chunksRead, t)
}

////////////////////////////////////////////////////////////////////////
// Test Function (Runs once before all tests)
////////////////////////////////////////////////////////////////////////

func Test(t *testing.T) {
	// Define flag set to run the tests.
	mountConfigFilePath := createConfigFile(9)
	flagSet := [][]string{
		{"--implicit-dirs=true", "--config-file=" + mountConfigFilePath},
		{"--implicit-dirs=false", "--config-file=" + mountConfigFilePath},
	}

	// Create storage client before running tests.
	var err error
	ts := &testStruct{}
	ctx, cancel := context.WithTimeout(ts.ctx, time.Minute*15)
	ts.storageClient, err = client.CreateStorageClient(ctx)
	if err != nil {
		log.Fatalf("client.CreateStorageClient: %v", err)
	}
	// Defer close storage client and release resources.
	defer func() {
		err := ts.storageClient.Close()
		if err != nil {
			t.Log("Failed to close storage client")
		}
		defer cancel()
	}()


	// Run tests.
	for _, flags := range flagSet {
		// Run tests without ro flag.
		ts.flags = flags
		test_setup.RunTests(t, ts)
		// Run tests with ro flag.
		ts.flags = append(flags, "--o=ro")
		test_setup.RunTests(t, ts)
	}
}
