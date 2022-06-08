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

package cloudfilesystem

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sealerio/sealer/common"
	"github.com/sealerio/sealer/pkg/env"
	"github.com/sealerio/sealer/pkg/runtime"
	v2 "github.com/sealerio/sealer/types/api/v2"
	"github.com/sealerio/sealer/utils/exec"
	"github.com/sealerio/sealer/utils/net"
	osi "github.com/sealerio/sealer/utils/os"
	"github.com/sealerio/sealer/utils/platform"
	"github.com/sealerio/sealer/utils/ssh"
	"github.com/sealerio/sealer/utils/strings"
	"github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"
)

const (
	RemoteNydusdInit = "cd %s && chmod +x *.sh && bash start.sh %s"
	RemoteNydusdStop = "if [ -f \"%[1]s\" ];then sh %[1]s;fi && rm -rf %s"
)

type nydusFileSystem struct {
}

func (n *nydusFileSystem) MountRootfs(cluster *v2.Cluster, hosts []string, initFlag bool) error {
	clusterRootfsDir := common.DefaultTheClusterRootfsDir(cluster.Name)
	//scp roofs to all Masters and Nodes,then do init.sh
	if err := mountNydusRootfs(hosts, clusterRootfsDir, cluster, initFlag); err != nil {
		return fmt.Errorf("mount rootfs failed %v", err)
	}

	return nil
}

func (n *nydusFileSystem) UnMountRootfs(cluster *v2.Cluster, hosts []string) error {
	var (
		nydusdFileDir     = common.DefaultTheClusterNydusdFileDir(cluster.Name)
		nydusdServerClean = filepath.Join(nydusdFileDir, "serverfile", "serverclean.sh")
	)
	//do clean.sh,then remove all Masters and Nodes roofs
	if err := unmountRootfs(hosts, cluster); err != nil {
		return err
	}

	if osi.IsFileExist(nydusdServerClean) {
		cleanCmd := fmt.Sprintf("sh %s", nydusdServerClean)
		_, err := exec.RunSimpleCmd(cleanCmd)
		if err != nil {
			return fmt.Errorf("failed to stop nydusdserver %v", err)
		}
	} else {
		logrus.Infof("%s not found", nydusdServerClean)
	}
	return nil
}

func mountNydusRootfs(ipList []string, target string, cluster *v2.Cluster, initFlag bool) error {
	clusterPlatform, err := ssh.GetClusterPlatform(cluster)
	if err != nil {
		return err
	}
	localIP, err := net.GetLocalIP(cluster.GetMaster0IP() + ":22")
	if err != nil {
		return fmt.Errorf("failed to get local address, %v", err)
	}
	var (
		nydusdfileSrc   = filepath.Join(platform.DefaultMountClusterImageDir(cluster.Name), "nydusdfile")
		nydusdFileDir   = common.DefaultTheClusterNydusdFileDir(cluster.Name)
		nydusdserverDir = filepath.Join(nydusdFileDir, "serverfile")
		nydusdfileCpCmd = fmt.Sprintf("rm -rf %s && cp -r %s %s", nydusdFileDir, nydusdfileSrc, nydusdFileDir)
		nydusdDir       = common.DefaultTheClusterNydusdDir(cluster.Name)
		nydusdInitCmd   = fmt.Sprintf(RemoteNydusdInit, nydusdDir, target)
		nydusdCleanCmd  = fmt.Sprintf(RemoteNydusdStop, filepath.Join(nydusdDir, "clean.sh"), nydusdDir)
		cleanCmd        = fmt.Sprintf("echo '%s' >> "+common.DefaultClusterClearBashFile, nydusdCleanCmd, cluster.Name)
		envProcessor    = env.NewEnvProcessor(cluster)
		config          = runtime.GetRegistryConfig(platform.DefaultMountClusterImageDir(cluster.Name), cluster.GetMaster0IP())
		initCmd         = fmt.Sprintf(RemoteChmod, target, config.Domain, config.Port)
	)
	_, err = exec.RunSimpleCmd(nydusdfileCpCmd)
	if err != nil {
		return fmt.Errorf("cp nydusdfile failed %v", err)
	}
	//dirs need be converted
	mountDirs := make(map[string]bool)
	dirlist := ""
	for _, IP := range ipList {
		ip := IP
		src := platform.GetMountClusterImagePlatformDir(cluster.Name, clusterPlatform[ip])
		if !mountDirs[src] {
			mountDirs[src] = true
			dirlist = dirlist + fmt.Sprintf(",%s", src)
			clientfileSrc := filepath.Join(src, "nydusdfile", "clientfile")
			clientfileDest := filepath.Join(nydusdFileDir, filepath.Base(src))
			nydusdCpCmd := fmt.Sprintf("cp -r %s %s", clientfileSrc, clientfileDest)
			_, err = exec.RunSimpleCmd(nydusdCpCmd)
			if err != nil {
				return fmt.Errorf("cp nydusdclinetfile failed %v", err)
			}
		}
	}
	startNydusdServer := fmt.Sprintf("cd %s && chmod +x serverstart.sh && ./serverstart.sh -d %s -i %s", nydusdserverDir, dirlist, localIP)
	//convert image and start nydusd http server
	_, err = exec.RunSimpleCmd(startNydusdServer)
	if err != nil {
		return fmt.Errorf("nydusdserver start fail %v", err)
	}
	logrus.Info("nydus images converted and nydusd http server started")

	eg, _ := errgroup.WithContext(context.Background())
	for _, IP := range ipList {
		ip := IP
		eg.Go(func() error {
			src := platform.GetMountClusterImagePlatformDir(cluster.Name, clusterPlatform[ip])
			src = filepath.Join(nydusdFileDir, filepath.Base(src))
			sshClient, err := ssh.GetHostSSHClient(ip, cluster)
			if err != nil {
				return fmt.Errorf("get host ssh client failed %v", err)
			}
			err = copyFiles(sshClient, ip, src, nydusdDir)
			if err != nil {
				return fmt.Errorf("scp nydusd failed %v", err)
			}
			if initFlag {
				err = sshClient.CmdAsync(ip, envProcessor.WrapperShell(ip, nydusdInitCmd))
				if err != nil {
					return fmt.Errorf("init nydusd failed %v", err)
				}
				err = sshClient.CmdAsync(ip, envProcessor.WrapperShell(ip, initCmd))
				if err != nil {
					return fmt.Errorf("exec init.sh failed %v", err)
				}
				err = sshClient.CmdAsync(ip, envProcessor.WrapperShell(ip, cleanCmd))
				if err != nil {
					return fmt.Errorf("echo nydusdcleancmd to clean.sh failed %v", err)
				}
			}
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		return err
	}
	if strings.NotIn(config.IP, ipList) {
		return nil
	}
	return nil
}

func NewNydusFileSystem() (Interface, error) {
	return &nydusFileSystem{}, nil
}
