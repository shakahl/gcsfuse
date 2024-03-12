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

package managed_folders

import (
	"fmt"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/creds_tests"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/operations"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/setup"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/test_setup"
	"log"
	"path"
	"testing"
)

// //////////////////////////////////////////////////////////////////////
// Boilerplate
// //////////////////////////////////////////////////////////////////////
const (
	AdminPermission           = "objectAdmin"
	IAMRoleForAdminPermission = "roles/storage.objectAdmin"
)

var (
	bucket         string
	testDir        string
	serviceAccount string
)

type managedFoldersAdminPermission struct {
}

func (s *managedFoldersAdminPermission) Setup(t *testing.T) {
}

func (s *managedFoldersAdminPermission) Teardown(t *testing.T) {
}

func (s *managedFoldersAdminPermission) TestListNonEmptyManagedFoldersWithAdminPermission(t *testing.T) {
	createDirectoryStructureForNonEmptyManagedFolders(t)

	listNonEmptyManagedFolders(t)

	cleanup(bucket, testDir, serviceAccount, IAMRoleForAdminPermission, t)
}

////////////////////////////////////////////////////////////////////////
// Test Function (Runs once before all tests)
////////////////////////////////////////////////////////////////////////

func TestManagedFolders_FolderAdminPermission(t *testing.T) {
	ts := &managedFoldersAdminPermission{}

	// Run tests for mountedDirectory only if --mountedDirectory  and --testBucket flag is set.
	if setup.AreBothMountedDirectoryAndTestBucketFlagsSet() {
		test_setup.RunTests(t, ts)
		return
	}

	// Fetch credentials and apply permission on bucket.
	serviceAccount, localKeyFilePath := creds_tests.CreateCredentials()
	creds_tests.ApplyPermissionToServiceAccount(serviceAccount, AdminPermission)
	// Revoke permission on bucket after unmounting and cleanup.
	defer creds_tests.RevokePermission(fmt.Sprintf("iam ch -d serviceAccount:%s:%s gs://%s", serviceAccount, AdminPermission, setup.TestBucket()))

	flags := []string{"--implicit-dirs", "--key-file=" + localKeyFilePath}

	if setup.OnlyDirMounted() != "" {
		operations.CreateManagedFoldersInBucket(onlyDirMounted, setup.TestBucket(), t)
		defer operations.DeleteManagedFoldersInBucket(onlyDirMounted, setup.TestBucket(), t)
	}
	setup.MountGCSFuseWithGivenMountFunc(flags, mountFunc)
	defer setup.UnmountGCSFuseAndDeleteLogFile(rootDir)
	setup.SetMntDir(mountDir)
	bucket, testDir = setup.GetBucketAndObjectBasedOnTypeOfMount(testDirNameForNonEmptyManagedFolder)

	// Run tests.
	log.Printf("Running tests with flags, bucket have admin permission and managed folder have nil permissions: %s", flags)
	test_setup.RunTests(t, ts)

	// Provide storage.objectViewer role to managed folders.
	providePermissionToManagedFolder(bucket, path.Join(testDir, ManagedFolder1), serviceAccount, IAMRoleForAdminPermission, t)
	providePermissionToManagedFolder(bucket, path.Join(testDir, ManagedFolder2), serviceAccount, IAMRoleForAdminPermission, t)

	log.Printf("Running tests with flags, bucket have admin permission and managed folder have admin permissions: %s", flags)
	test_setup.RunTests(t, ts)

	// Revoke permission on bucket after unmounting and cleanup.
	creds_tests.RevokePermission(fmt.Sprintf("iam ch -d serviceAccount:%s:%s gs://%s", serviceAccount, AdminPermission, setup.TestBucket()))
	creds_tests.ApplyPermissionToServiceAccount(serviceAccount, ViewPermission)
	defer creds_tests.RevokePermission(fmt.Sprintf("iam ch -d serviceAccount:%s:%s gs://%s", serviceAccount, ViewPermission, setup.TestBucket()))
	// Run tests.
	log.Printf("Running tests with flags, bucket have view permission and managed folder have admin permissions: %s", flags)
	test_setup.RunTests(t, ts)
}
