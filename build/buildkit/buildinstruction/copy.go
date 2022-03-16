// Copyright © 2021 Alibaba Group Holding Ltd.
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

package buildinstruction

import (
	"fmt"
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/logger"
	"github.com/alibaba/sealer/pkg/image/cache"
	"github.com/alibaba/sealer/pkg/image/store"
	v1 "github.com/alibaba/sealer/types/api/v1"
	"github.com/alibaba/sealer/utils"
	"github.com/alibaba/sealer/utils/collector"
	"github.com/opencontainers/go-digest"
)

type CopyInstruction struct {
	src       string
	dest      string
	rawLayer  v1.Layer
	fs        store.Backend
	collector collector.Collector
}

func (c CopyInstruction) Exec(execContext ExecContext) (out Out, err error) {
	var (
		hitCache bool
		chainID  cache.ChainID
		cacheID  digest.Digest
		layerID  digest.Digest
	)
	defer func() {
		out.ContinueCache = hitCache
		out.ParentID = chainID
	}()

	if !isRemoteSource(c.src) {
		cacheID, err = GenerateSourceFilesDigest(execContext.BuildContext, c.src)
		if err != nil {
			logger.Warn("failed to generate src digest,discard cache,%s", err)
		}

		if execContext.ContinueCache {
			hitCache, layerID, chainID = tryCache(execContext.ParentID, c.rawLayer, execContext.CacheSvc, execContext.Prober, cacheID)
			// we hit the cache, so we will reuse the layerID layer.
			if hitCache {
				// update chanid as parentid via defer
				out.LayerID = layerID
				return out, nil
			}
		}
	}

	tmp, err := utils.MkTmpdir()
	if err != nil {
		return out, fmt.Errorf("failed to create tmp dir %s:%v", tmp, err)
	}

	err = c.collector.Collect(execContext.BuildContext, c.src, filepath.Join(tmp, c.dest))
	if err != nil {
		return out, fmt.Errorf("failed to collect files to temp dir %s, err: %v", tmp, err)
	}
	// if we come here, its new layer need set cache id .
	layerID, err = execContext.LayerStore.RegisterLayerForBuilder(tmp)
	if err != nil {
		return out, fmt.Errorf("failed to register copy layer, err: %v", err)
	}

	if setErr := c.SetCacheID(layerID, cacheID.String()); setErr != nil {
		logger.Warn("set cache failed layer: %v, err: %v", c.rawLayer, err)
	}

	out.LayerID = layerID
	return out, nil
}

// SetCacheID This function only has meaning for copy layers
func (c CopyInstruction) SetCacheID(layerID digest.Digest, cID string) error {
	return c.fs.SetMetadata(layerID, common.CacheID, []byte(cID))
}

func NewCopyInstruction(ctx InstructionContext) (*CopyInstruction, error) {
	fs, err := store.NewFSStoreBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to init store backend, err: %s", err)
	}
	src, dest := ParseCopyLayerContent(ctx.CurrentLayer.Value)
	c, err := collector.NewCollector(src)
	if err != nil {
		return nil, fmt.Errorf("failed to init copy Collector, err: %s", err)
	}

	return &CopyInstruction{
		fs:        fs,
		rawLayer:  *ctx.CurrentLayer,
		src:       src,
		dest:      dest,
		collector: c,
	}, nil
}
