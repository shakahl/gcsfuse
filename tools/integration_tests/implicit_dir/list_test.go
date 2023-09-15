// Copyright 2023 Google Inc. All Rights Reserved.
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

package implicit_dir_test

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/setup"
	"github.com/googlecloudplatform/gcsfuse/tools/integration_tests/util/setup/implicit_and_explicit_dir_setup"
)

const DirectoryForImplicitDirListTesting = "directoryForImplicitDirListTesting"

func TestListImplicitObjectsFromBucket(t *testing.T) {
	// Clean the test Directory before running test.
	setup.PreTestSetup(DirectoryForImplicitDirListTesting)

	// Clean the test Directory after running test.
	testDir := path.Join(setup.MntDir(), DirectoryForImplicitDirListTesting)
	defer setup.CleanUpDir(testDir)

	// Directory Structure
	// testBucket/directoryForImplicitDirListTesting/implicitDirectory                                                  -- Dir
	// testBucket/directoryForImplicitDirListTesting/implicitDirectory/fileInImplicitDir1                               -- File
	// testBucket/directoryForImplicitDirListTesting/implicitDirectory/implicitSubDirectory                             -- Dir
	// testBucket/directoryForImplicitDirListTesting/implicitDirectory/implicitSubDirectory/fileInImplicitDir2          -- File
	// testBucket/directoryForImplicitDirListTesting/explicitDirectory                                                  -- Dir
	// testBucket/directoryForImplicitDirListTesting/explicitFile                                                       -- File
	// testBucket/directoryForImplicitDirListTesting/explicitDirectory/fileInExplicitDir1                               -- File
	// testBucket/directoryForImplicitDirListTesting/explicitDirectory/fileInExplicitDir2                               -- File

	implicit_and_explicit_dir_setup.CreateImplicitDirectoryStructure(DirectoryForImplicitDirListTesting)
	implicit_and_explicit_dir_setup.CreateExplicitDirectoryStructure(DirectoryForImplicitDirListTesting, t)

	err := filepath.WalkDir(testDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// The object type is not directory.
		if !dir.IsDir() {
			return nil
		}

		objs, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		// Check if mntDir has correct objects.
		if path == testDir {
			// numberOfObjects - 3
			if len(objs) != implicit_and_explicit_dir_setup.NumberOfTotalObjects {
				t.Errorf("Incorrect number of objects in the test directory Expected:%d Actual:%d", implicit_and_explicit_dir_setup.NumberOfTotalObjects, len(objs))
			}

			// testBucket/directoryForImplicitDirListTesting/explicitDir     -- Dir
			if objs[0].Name() != implicit_and_explicit_dir_setup.ExplicitDirectory || objs[0].IsDir() != true {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.ExplicitDirectory, objs[0].Name())
			}
			// testBucket/directoryForImplicitDirListTesting/explicitFile    -- File
			if objs[1].Name() != implicit_and_explicit_dir_setup.ExplicitFile || objs[1].IsDir() != false {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.ExplicitFile, objs[1].Name())
			}

			// testBucket/directoryForImplicitDirListTesting/implicitDir     -- Dir
			if objs[2].Name() != implicit_and_explicit_dir_setup.ImplicitDirectory || objs[2].IsDir() != true {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.ImplicitDirectory, objs[2].Name())
			}
		}

		// Check if explictDir directory has correct data.
		if dir.IsDir() && dir.Name() == implicit_and_explicit_dir_setup.ExplicitDirectory {
			// numberOfObjects - 2
			if len(objs) != implicit_and_explicit_dir_setup.NumberOfFilesInExplicitDirectory {
				t.Errorf("Incorrect number of objects in the explicitDirectory Expected:%d Actual:%d", implicit_and_explicit_dir_setup.NumberOfFilesInExplicitDirectory, len(objs))
			}

			// testBucket/directoryForImplicitDirListTesting/explicitDir/fileInExplicitDir1   -- File
			if objs[0].Name() != implicit_and_explicit_dir_setup.FirstFileInExplicitDirectory || objs[0].IsDir() != false {
				t.Errorf("Listed incorrect object  Expected:%s Actual:%s", implicit_and_explicit_dir_setup.FirstFileInExplicitDirectory, objs[0].Name())
			}

			// testBucket/directoryForImplicitDirListTesting/explicitDir/fileInExplicitDir2    -- File
			if objs[1].Name() != implicit_and_explicit_dir_setup.SecondFileInExplicitDirectory || objs[1].IsDir() != false {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.SecondFileInExplicitDirectory, objs[1].Name())
			}
			return nil
		}

		// Check if implicitDir directory has correct data.
		if dir.IsDir() && dir.Name() == implicit_and_explicit_dir_setup.ImplicitDirectory {
			// numberOfObjects - 2
			if len(objs) != implicit_and_explicit_dir_setup.NumberOfFilesInImplicitDirectory {
				t.Errorf("Incorrect number of objects in the implicitDirectory Expected:%d Actual:%d", implicit_and_explicit_dir_setup.NumberOfFilesInImplicitDirectory, len(objs))
			}

			// testBucket/directoryForImplicitDirListTesting/implicitDir/fileInImplicitDir1  -- File
			if objs[0].Name() != implicit_and_explicit_dir_setup.FileInImplicitDirectory || objs[0].IsDir() != false {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.FileInImplicitDirectory, objs[0].Name())
			}
			// testBucket/directoryForImplicitDirListTesting/implicitDir/implicitSubDirectory  -- Dir
			if objs[1].Name() != implicit_and_explicit_dir_setup.ImplicitSubDirectory || objs[1].IsDir() != true {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.ImplicitSubDirectory, objs[1].Name())
			}
			return nil
		}

		// Check if implicit sub directory has correct data.
		if dir.IsDir() && dir.Name() == implicit_and_explicit_dir_setup.ImplicitSubDirectory {
			// numberOfObjects - 1
			if len(objs) != implicit_and_explicit_dir_setup.NumberOfFilesInImplicitSubDirectory {
				t.Errorf("Incorrect number of objects in the implicitSubDirectoryt Expected:%d Actual:%d", implicit_and_explicit_dir_setup.NumberOfFilesInImplicitSubDirectory, len(objs))
			}

			// testBucket/directoryForImplicitDirListTesting/implicitDir/implicitSubDir/fileInImplicitDir2   -- File
			if objs[0].Name() != implicit_and_explicit_dir_setup.FileInImplicitSubDirectory || objs[0].IsDir() != false {
				t.Errorf("Listed incorrect object Expected:%s Actual:%s", implicit_and_explicit_dir_setup.FileInImplicitSubDirectory, objs[0].Name())
			}
			return nil
		}
		return nil
	})
	if err != nil {
		t.Errorf("error walking the path : %v\n", err)
		return
	}
}
