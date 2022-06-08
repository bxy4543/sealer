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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/sealerio/sealer/pkg/image"
	"github.com/sealerio/sealer/pkg/image/utils"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push ClusterImage to remote registry",
	// TODO: add long description.
	Long:    "",
	Example: `sealer push registry.cn-qingdao.aliyuncs.com/sealer-io/my-kubernetes-cluster-with-dashboard:latest`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imgsvc, err := image.NewImageService()
		if err != nil {
			return err
		}

		return imgsvc.Push(args[0])

	},
	ValidArgsFunction: utils.ImageListFuncForCompletion,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
