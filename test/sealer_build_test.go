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
	"os"
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/test/suites/apply"
	"github.com/alibaba/sealer/test/suites/image"
	"github.com/alibaba/sealer/test/suites/registry"

	"github.com/alibaba/sealer/test/suites/build"
	"github.com/alibaba/sealer/test/testhelper"
	"github.com/alibaba/sealer/test/testhelper/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("sealer build", func() {
	Context("test build args", func() {
		Context("build kube file", func() {
			Context("testing the abnormal scenario ", func() {
				var tempFile string
				BeforeEach(func() {
					tempFile = testhelper.CreateTempFile()
				})

				AfterEach(func() {
					testhelper.RemoveTempFile(tempFile)
				})
				It("specific the base rootfs that do not exist", func() {
					//base rootfs not exist
					imageName := build.GetImageNameTemplate("no_rootfs")
					err := testhelper.WriteFile(tempFile, []byte("from abc\ncopy . ."))
					Expect(err).NotTo(HaveOccurred())
					cmd := build.NewArgsOfBuild().
						SetKubeFile(tempFile).
						SetImageName(imageName).
						SetContext(".").
						SetBuildType(common.LocalBuild).
						Build()
					sess, err := testhelper.Start(cmd)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(2))
					Expect(build.CheckIsImageExist(imageName)).ShouldNot(BeTrue())
				})

				It("copy src that do not exist", func() {
					//copy: copy src not exist;
					imageName := build.GetImageNameTemplate("no_src_copy")
					err := testhelper.WriteFile(tempFile, []byte("from sealer-io/kubernetes:v1.19.9\ncopy abc123 ."))
					Expect(err).NotTo(HaveOccurred())
					cmd := build.NewArgsOfBuild().
						SetKubeFile(tempFile).
						SetImageName(imageName).
						SetContext(".").
						SetBuildType(common.LocalBuild).
						Build()
					sess, err := testhelper.Start(cmd)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(2))
					Expect(build.CheckIsImageExist(imageName)).ShouldNot(BeTrue())
				})

				It("exec cmd that do not exist in system", func() {
					//run&cmd: exec cmd not exist
					imageName := build.GetImageNameTemplate("no_cmd_run")
					err := testhelper.WriteFile(tempFile, []byte("from sealer-io/kubernetes:v1.19.9\ncmd abc ."))
					Expect(err).NotTo(HaveOccurred())
					cmd := build.NewArgsOfBuild().
						SetKubeFile(tempFile).
						SetImageName(imageName).
						SetContext(".").
						SetBuildType(common.LocalBuild).
						Build()
					sess, err := testhelper.Start(cmd)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(2))
					Expect(build.CheckIsImageExist(imageName)).ShouldNot(BeTrue())
				})
			})

			Context("testing the content of kube file", func() {
				Context("testing local build scenario", func() {
					err := os.Chdir(filepath.Join(build.GetFixtures(), build.GetLocalBuildDir()))
					Expect(err).NotTo(HaveOccurred())

					BeforeEach(func() {
						registry.Login()
					})
					AfterEach(func() {
						registry.Logout()
					})

					It("with all build instruct", func() {
						imageName := build.GetImageNameTemplate("all_instruct")
						cmd := build.NewArgsOfBuild().
							SetKubeFile("Kubefile").
							SetImageName(imageName).
							SetContext(".").
							SetBuildType(settings.LocalBuild).
							Build()
						sess, err := testhelper.Start(cmd)
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))
						// check: sealer images whether image exist
						Expect(build.CheckIsImageExist(imageName)).Should(BeTrue())
						Expect(build.CheckClusterFile(imageName)).Should(BeTrue())
					})

					It("only copy instruct", func() {
						imageName := build.GetImageNameTemplate("only_copy")
						cmd := build.NewArgsOfBuild().
							SetKubeFile("Kubefile_only_copy").
							SetImageName(imageName).
							SetContext(".").
							SetBuildType(settings.LocalBuild).
							Build()
						sess, err := testhelper.Start(cmd)
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))
						// check: sealer images whether image exist
						Expect(build.CheckIsImageExist(imageName)).Should(BeTrue())
						Expect(build.CheckClusterFile(imageName)).Should(BeTrue())
					})

				})
				Context("testing cloud build scenario", func() {
					BeforeEach(func() {
						registry.Login()
						err := os.Chdir(filepath.Join("..", build.GetCloudBuildDir()))
						Expect(err).NotTo(HaveOccurred())
					})
					AfterEach(func() {
						registry.Logout()
					})

					It("with all build instruct", func() {
						imageName := build.GetTestImageName()
						cmd := build.NewArgsOfBuild().
							SetKubeFile("Kubefile").
							SetImageName(imageName).
							SetContext(".").
							SetBuildType("cloud").
							Build()
						sess, err := testhelper.Start(cmd)
						defer func() {
							if testhelper.IsFileExist(settings.TMPClusterFile) {
								cluster := apply.LoadClusterFileFromDisk(settings.TMPClusterFile)
								apply.CleanUpAliCloudInfra(cluster)
								testhelper.DeleteFileLocally(settings.TMPClusterFile)
							}
						}()
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
						// check: need to pull build image and check whether image exist
						image.DoImageOps(settings.SubCmdPullOfSealer, imageName)
						Expect(build.CheckIsImageExist(imageName)).Should(BeTrue())
						Expect(build.CheckClusterFile(imageName)).Should(BeTrue())
					})

				})

			})

		})
	})
})
