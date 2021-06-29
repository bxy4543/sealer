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

package test

import (
	"strconv"
	"strings"

	"github.com/alibaba/sealer/test/suites/apply"
	"github.com/alibaba/sealer/test/testhelper"
	"github.com/alibaba/sealer/test/testhelper/settings"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sealer join", func() {
	/*	Context("join cloud ", func() {
		AfterEach(func() {
			apply.DeleteClusterByFile(settings.GetClusterWorkClusterfile(settings.ClusterNameForRun))
		})

		It("exec sealer join", func() {
			//init cluster 1 master 1 node
			nodeNum := 0
			apply.SealerRun(strconv.Itoa(1), strconv.Itoa(1))
			nodeNum = 2
			apply.CheckNodeNumLocally(nodeNum)

			By("join master and node", func() {
				//join 2 master and 1 node
				apply.SealerJoin(strconv.Itoa(2), strconv.Itoa(1))
				nodeNum += 3
				apply.CheckNodeNumLocally(nodeNum)
			})
			By("join master only", func() {
				//join 2 master
				apply.SealerJoin(strconv.Itoa(2), "")
				nodeNum += 2
				apply.CheckNodeNumLocally(nodeNum)
			})
			By("join node only", func() {
				//join 1 node
				apply.SealerJoin("", strconv.Itoa(1))
				nodeNum++
				apply.CheckNodeNumLocally(nodeNum)
			})
		})
	})*/

	Context("join bareMetal", func() {
		var tempFile string
		BeforeEach(func() {
			tempFile = testhelper.CreateTempFile()
		})

		AfterEach(func() {
			testhelper.RemoveTempFile(tempFile)
		})

		It("bareMetal join", func() {
			rawClusterFilePath := apply.GetRawClusterFilePath()
			rawCluster := apply.LoadClusterFileFromDisk(rawClusterFilePath)
			By("start to prepare infra")
			usedCluster := apply.CreateAliCloudInfraAndSave(rawCluster, tempFile)
			//defer to delete cluster
			defer func() {
				apply.CleanUpAliCloudInfra(usedCluster)
			}()
			sshClient := testhelper.NewSSHClientByCluster(usedCluster)
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())

			By("start to init cluster", func() {
				apply.SendAndApplyCluster(sshClient, tempFile)
				apply.CheckNodeNumWithSSH(sshClient, 2)
			})

			By("start to join master and node", func() {
				usedCluster.Spec.Masters.Count = strconv.Itoa(3)
				usedCluster.Spec.Nodes.Count = strconv.Itoa(2)
				//create infra
				usedCluster = apply.CreateAliCloudInfraAndSave(usedCluster, tempFile)
				joinMasters := strings.Join(usedCluster.Spec.Masters.IPList[1:], ",")
				joinNodes := strings.Join(usedCluster.Spec.Nodes.IPList[1:], ",")
				//sealer join master and node
				apply.SendAndJoinCluster(sshClient, tempFile, joinMasters, joinNodes)
				//add 3 masters and 2 nodes
				apply.CheckNodeNumWithSSH(sshClient, 5)
			})

			By("join master only", func() {
				usedCluster.Spec.Masters.Count = strconv.Itoa(5)
				//create infra
				usedCluster = apply.CreateAliCloudInfraAndSave(usedCluster, tempFile)
				joinMasters := strings.Join(usedCluster.Spec.Masters.IPList[3:], ",")
				apply.SendAndJoinCluster(sshClient, tempFile, joinMasters, "")
				//add 2 masters
				apply.CheckNodeNumWithSSH(sshClient, 7)
			})

			By("join node only", func() {
				usedCluster.Spec.Nodes.Count = strconv.Itoa(3)
				//create infra
				usedCluster = apply.CreateAliCloudInfraAndSave(usedCluster, tempFile)
				joinNodes := strings.Join(usedCluster.Spec.Nodes.IPList[2:], ",")
				apply.SendAndJoinCluster(sshClient, tempFile, "", joinNodes)
				//add 1 node
				apply.CheckNodeNumWithSSH(sshClient, 8)
			})
		})
	})

})
