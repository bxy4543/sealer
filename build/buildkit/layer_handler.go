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

package buildkit

import (
	"strings"

	"github.com/alibaba/sealer/build/buildkit/buildlayer"
	"github.com/alibaba/sealer/common"
	v1 "github.com/alibaba/sealer/types/api/v1"
)

// LayerValueExchange :init different layer handler to exchanging due to the layer content
func LayerValueExchange(layer *v1.Layer) buildlayer.LayerHandler {
	var layerParser buildlayer.LayerCopy
	if layer.Type == common.COPYCOMMAND {
		layerParser = parseCopyLayerValue(layer.Value)
	}

	switch layerParser.HandlerType {
	// imageList;yaml,chart
	case ImageListHandler:
		return buildlayer.NewImageListHandler(layerParser)
	case YamlHandler:
		return buildlayer.NewYamlHandler(layerParser)
	case ChartHandler:
		return buildlayer.NewChartHandler(layerParser)
	}
	return nil
}

func parseCopyLayerValue(layerValue string) buildlayer.LayerCopy {
	//COPY imageList manifests
	//COPY cc charts
	//COPY recommended.yaml manifests
	//COPY nginx.tar images

	lc := buildlayer.LayerCopy{
		Src:  strings.Fields(layerValue)[0],
		Dest: strings.TrimPrefix(strings.Fields(layerValue)[1], "./"),
	}
	if lc.Dest == IsCopyToManifests {
		if lc.Src == ImageList {
			lc.HandlerType = ImageListHandler
		}
		if strings.HasSuffix(lc.Src, ".yaml") || strings.HasSuffix(lc.Src, ".yml") {
			lc.HandlerType = YamlHandler
		}
		return lc
	}

	if lc.Dest == IsCopyToChart {
		lc.HandlerType = ChartHandler
		return lc
	}

	if lc.Dest == IsCopyOfflineImage {
		lc.HandlerType = OfflineImageHandler
		return lc
	}

	return lc
}
