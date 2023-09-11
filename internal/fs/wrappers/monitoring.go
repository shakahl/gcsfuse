// Copyright 2020 Google Inc. All Rights Reserved.
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

package wrappers

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/googlecloudplatform/gcsfuse/internal/logger"
	"github.com/googlecloudplatform/gcsfuse/internal/monitor/tags"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	opsCount      = stats.Int64("fs/ops_count", "The number of ops processed by the file system.", stats.UnitDimensionless)
	opsLatency    = stats.Float64("fs/ops_latency", "The latency of a file system operation.", stats.UnitMilliseconds)
	opsErrorCount = stats.Int64("fs/ops_error_count", "The number of errors generated by file system operation.", stats.UnitDimensionless)
)

// Initialize the metrics.
func init() {

	// Register the view.
	if err := view.Register(
		&view.View{
			Name:        "fs/ops_count",
			Measure:     opsCount,
			Description: "The cumulative number of ops processed by the file system.",
			Aggregation: view.Sum(),
			TagKeys:     []tag.Key{tags.FSOp},
		},
		&view.View{
			Name:        "fs/ops_error_count",
			Measure:     opsErrorCount,
			Description: "The cumulative number of errors generated by file system operations",
			Aggregation: view.Sum(),
			TagKeys:     []tag.Key{tags.FSOp, tags.FSError},
		},
		&view.View{
			Name:        "fs/ops_latency",
			Measure:     opsLatency,
			Description: "The cumulative distribution of file system operation latencies",
			Aggregation: ochttp.DefaultLatencyDistribution,
			TagKeys:     []tag.Key{tags.FSOp},
		}); err != nil {
		fmt.Printf("Failed to register metrics for the file system: %v\n", err)
	}
}

// fsErrStr maps an error to a error string. Uncommon errors are aggregated to
// reduce the cardinality of the fs error to save the monitoring cost.
func fsErrStr(err error) string {
	if err == nil {
		return ""
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno.Error()
	}
	return DefaultFSError.Error()
}

// Records file system operation count, failed operation count and the operation latency.
func recordOp(ctx context.Context, method string, start time.Time, fsErr error) {

	// Recording opCount.
	if err := stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Upsert(tags.FSOp, method),
		},
		opsCount.M(1),
	); err != nil {
		// Error in recording opCount.
		fmt.Printf("Cannot record file system op: %v", err)
	}

	// Recording opErrorCount.
	if fsErr != nil {
		if err := stats.RecordWithTags(
			ctx,
			[]tag.Mutator{
				tag.Upsert(tags.FSOp, method),
				tag.Upsert(tags.FSError, fsErrStr(fsErr)),
			},
			opsErrorCount.M(1),
		); err != nil {
			// Error in recording opErrorCount.
			logger.Errorf("Cannot record error count of the file system failed operations: %v", err)
		}
	}

	// Recording opLatency.
	latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
	if err := stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Upsert(tags.FSOp, method),
		},
		opsLatency.M(latencyMs),
	); err != nil {
		// Error in opLatency.
		logger.Errorf("Cannot record file system operation latency: %v", err)
	}
}

// WithMonitoring takes a FileSystem, returns a FileSystem with monitoring
// on the counts of requests per API.
func WithMonitoring(fs fuseutil.FileSystem) fuseutil.FileSystem {
	return &monitoring{
		wrapped: fs,
	}
}

type monitoring struct {
	wrapped fuseutil.FileSystem
}

func (fs *monitoring) Destroy() {
	fs.wrapped.Destroy()
}

func (fs *monitoring) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	startTime := time.Now()
	err := fs.wrapped.StatFS(ctx, op)
	recordOp(ctx, "StatFS", startTime, err)
	return err
}

func (fs *monitoring) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	startTime := time.Now()
	err := fs.wrapped.LookUpInode(ctx, op)
	recordOp(ctx, "LookUpInode", startTime, err)
	return err
}

func (fs *monitoring) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	startTime := time.Now()
	err := fs.wrapped.GetInodeAttributes(ctx, op)
	recordOp(ctx, "GetInodeAttributes", startTime, err)
	return err
}

func (fs *monitoring) SetInodeAttributes(
	ctx context.Context,
	op *fuseops.SetInodeAttributesOp) error {
	startTime := time.Now()
	err := fs.wrapped.SetInodeAttributes(ctx, op)
	recordOp(ctx, "SetInodeAttributes", startTime, err)
	return err
}

func (fs *monitoring) ForgetInode(
	ctx context.Context,
	op *fuseops.ForgetInodeOp) error {
	startTime := time.Now()
	err := fs.wrapped.ForgetInode(ctx, op)
	recordOp(ctx, "ForgetInode", startTime, err)
	return err
}

func (fs *monitoring) BatchForget(
	ctx context.Context,
	op *fuseops.BatchForgetOp) error {
	startTime := time.Now()
	err := fs.wrapped.BatchForget(ctx, op)
	recordOp(ctx, "BatchForget", startTime, err)
	return err
}

func (fs *monitoring) MkDir(
	ctx context.Context,
	op *fuseops.MkDirOp) error {
	startTime := time.Now()
	err := fs.wrapped.MkDir(ctx, op)
	recordOp(ctx, "MkDir", startTime, err)
	return err
}

func (fs *monitoring) MkNode(
	ctx context.Context,
	op *fuseops.MkNodeOp) error {
	startTime := time.Now()
	err := fs.wrapped.MkNode(ctx, op)
	recordOp(ctx, "MkNode", startTime, err)
	return err
}

func (fs *monitoring) CreateFile(
	ctx context.Context,
	op *fuseops.CreateFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.CreateFile(ctx, op)
	recordOp(ctx, "CreateFile", startTime, err)
	return err
}

func (fs *monitoring) CreateLink(
	ctx context.Context,
	op *fuseops.CreateLinkOp) error {
	startTime := time.Now()
	err := fs.wrapped.CreateLink(ctx, op)
	recordOp(ctx, "CreateLink", startTime, err)
	return err
}

func (fs *monitoring) CreateSymlink(
	ctx context.Context,
	op *fuseops.CreateSymlinkOp) error {
	startTime := time.Now()
	err := fs.wrapped.CreateSymlink(ctx, op)
	recordOp(ctx, "CreateSymlink", startTime, err)
	return err
}

func (fs *monitoring) Rename(
	ctx context.Context,
	op *fuseops.RenameOp) error {
	startTime := time.Now()
	err := fs.wrapped.Rename(ctx, op)
	recordOp(ctx, "Rename", startTime, err)
	return err
}

func (fs *monitoring) RmDir(
	ctx context.Context,
	op *fuseops.RmDirOp) error {
	startTime := time.Now()
	err := fs.wrapped.RmDir(ctx, op)
	recordOp(ctx, "RmDir", startTime, err)
	return err
}

func (fs *monitoring) Unlink(
	ctx context.Context,
	op *fuseops.UnlinkOp) error {
	startTime := time.Now()
	err := fs.wrapped.Unlink(ctx, op)
	recordOp(ctx, "Unlink", startTime, err)
	return err
}

func (fs *monitoring) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	startTime := time.Now()
	err := fs.wrapped.OpenDir(ctx, op)
	recordOp(ctx, "OpenDir", startTime, err)
	return err
}

func (fs *monitoring) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	startTime := time.Now()
	err := fs.wrapped.ReadDir(ctx, op)
	recordOp(ctx, "ReadDir", startTime, err)
	return err
}

func (fs *monitoring) ReleaseDirHandle(
	ctx context.Context,
	op *fuseops.ReleaseDirHandleOp) error {
	startTime := time.Now()
	err := fs.wrapped.ReleaseDirHandle(ctx, op)
	recordOp(ctx, "ReleaseDirHandle", startTime, err)
	return err
}

func (fs *monitoring) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.OpenFile(ctx, op)
	recordOp(ctx, "OpenFile", startTime, err)
	return err
}

func (fs *monitoring) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.ReadFile(ctx, op)
	recordOp(ctx, "ReadFile", startTime, err)
	return err
}

func (fs *monitoring) WriteFile(
	ctx context.Context,
	op *fuseops.WriteFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.WriteFile(ctx, op)
	recordOp(ctx, "WriteFile", startTime, err)
	return err
}

func (fs *monitoring) SyncFile(
	ctx context.Context,
	op *fuseops.SyncFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.SyncFile(ctx, op)
	recordOp(ctx, "SyncFile", startTime, err)
	return err
}

func (fs *monitoring) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) error {
	startTime := time.Now()
	err := fs.wrapped.FlushFile(ctx, op)
	recordOp(ctx, "FlushFile", startTime, err)
	return err
}

func (fs *monitoring) ReleaseFileHandle(
	ctx context.Context,
	op *fuseops.ReleaseFileHandleOp) error {
	startTime := time.Now()
	err := fs.wrapped.ReleaseFileHandle(ctx, op)
	recordOp(ctx, "ReleaseFileHandle", startTime, err)
	return err
}

func (fs *monitoring) ReadSymlink(
	ctx context.Context,
	op *fuseops.ReadSymlinkOp) error {
	startTime := time.Now()
	err := fs.wrapped.ReadSymlink(ctx, op)
	recordOp(ctx, "ReadSymlink", startTime, err)
	return err
}

func (fs *monitoring) RemoveXattr(
	ctx context.Context,
	op *fuseops.RemoveXattrOp) error {
	startTime := time.Now()
	err := fs.wrapped.RemoveXattr(ctx, op)
	recordOp(ctx, "RemoveXattr", startTime, err)
	return err
}

func (fs *monitoring) GetXattr(
	ctx context.Context,
	op *fuseops.GetXattrOp) error {
	startTime := time.Now()
	err := fs.wrapped.GetXattr(ctx, op)
	recordOp(ctx, "GetXattr", startTime, err)
	return err
}

func (fs *monitoring) ListXattr(
	ctx context.Context,
	op *fuseops.ListXattrOp) error {
	startTime := time.Now()
	err := fs.wrapped.ListXattr(ctx, op)
	recordOp(ctx, "ListXattr", startTime, err)
	return err
}

func (fs *monitoring) SetXattr(
	ctx context.Context,
	op *fuseops.SetXattrOp) error {
	startTime := time.Now()
	err := fs.wrapped.SetXattr(ctx, op)
	recordOp(ctx, "SetXattr", startTime, err)
	return err
}

func (fs *monitoring) Fallocate(
	ctx context.Context,
	op *fuseops.FallocateOp) error {
	startTime := time.Now()
	err := fs.wrapped.Fallocate(ctx, op)
	recordOp(ctx, "Fallocate", startTime, err)
	return err
}
