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

	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"

	"github.com/sealerio/sealer/common"
	"github.com/sealerio/sealer/pkg/image/cache"
	v1 "github.com/sealerio/sealer/types/api/v1"
	"github.com/sealerio/sealer/utils/exec"
	"github.com/sealerio/sealer/utils/mount"
	"github.com/sealerio/sealer/utils/os"
)

type CmdInstruction struct {
	cmdValue string
	rawLayer v1.Layer
	mounter  mount.Service
	ex       *shell.Lex
}

func (c CmdInstruction) Exec(execContext ExecContext) (out Out, err error) {
	var (
		hitCache bool
		chainID  cache.ChainID
		layerID  digest.Digest
	)
	defer func() {
		out.ContinueCache = hitCache
		out.ParentID = chainID
	}()

	if execContext.ContinueCache {
		hitCache, layerID, chainID = tryCache(execContext.ParentID, c.rawLayer, execContext.CacheSvc, execContext.Prober, "")
		if hitCache {
			out.LayerID = layerID
			return out, nil
		}
	}

	err = c.mounter.TempMount()
	if err != nil {
		return out, err
	}
	defer c.mounter.CleanUp()

	err = os.SetRootfsBinToSystemEnv(c.mounter.GetMountTarget())
	if err != nil {
		return out, fmt.Errorf("failed to set temp rootfs %s to system $PATH : %v", c.mounter.GetMountTarget(), err)
	}

	// if no variable at cmd value,nothing will change.
	// if no build args is matched at cmd value,then the variable will be null.
	cmdline, err := c.ex.ProcessWordWithMap(c.cmdValue, execContext.BuildArgs)
	if err != nil {
		return out, fmt.Errorf("failed to render build args: %v", err)
	}

	cmd := fmt.Sprintf(common.CdAndExecCmd, c.mounter.GetMountTarget(), cmdline)
	output, err := exec.RunSimpleCmd(cmd)
	logrus.Info(output)

	if err != nil {
		return out, fmt.Errorf("failed to exec %s, err: %v", cmd, err)
	}

	// cmd do not contain layer ,so no need to calculate layer
	if c.rawLayer.Type == common.CMDCOMMAND {
		return out, nil
	}

	out.LayerID, err = execContext.LayerStore.RegisterLayerForBuilder(c.mounter.GetMountUpper())
	return out, err
}

func NewCmdInstruction(ctx InstructionContext) (*CmdInstruction, error) {
	lowerLayers := GetBaseLayersPath(ctx.BaseLayers)
	mountService, err := mount.NewMountService("", "", lowerLayers)
	if err != nil {
		return nil, err
	}

	return &CmdInstruction{
		mounter:  mountService,
		cmdValue: ctx.CurrentLayer.Value,
		rawLayer: *ctx.CurrentLayer,
		ex:       shell.NewLex('\\'),
	}, nil
}
