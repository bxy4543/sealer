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

package image

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/alibaba/sealer/test/testhelper/settings"

	"github.com/alibaba/sealer/test/testhelper"
)

func DoImageOps(action, imageName string) {
	cmd := ""
	switch action {
	case "pull":
		cmd = fmt.Sprintf("%s pull %s", settings.DefaultSealerBin, imageName)
	case "push":
		cmd = fmt.Sprintf("%s push %s", settings.DefaultSealerBin, imageName)
	case "rmi":
		cmd = fmt.Sprintf("%s rmi %s", settings.DefaultSealerBin, imageName)
	case "run":
		cmd = fmt.Sprintf("%s run %s", settings.DefaultSealerBin, imageName)
	case "images":
		cmd = fmt.Sprintf("%s images", settings.DefaultSealerBin)
	}

	testhelper.RunCmdAndCheckResult(cmd, 0)
}
func TagImages(oldName, newName string) {
	cmd := fmt.Sprintf("%s tag %s %s", settings.DefaultSealerBin, oldName, newName)
	testhelper.RunCmdAndCheckResult(cmd, 0)
}

func GetEnvDirMd5() string {
	cmd := fmt.Sprintf("sudo -E find %s -type f -print0|xargs -0 sudo md5sum|cut -d\" \" -f1|md5sum|cut -d\" \" -f1\n", filepath.Dir(common.DefaultImageRootDir))
	dirMd5, err := utils.RunSimpleCmd(cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = io.WriteString(ginkgo.GinkgoWriter, cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return dirMd5
}
