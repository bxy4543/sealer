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

package plugin

/*
func TestDumperPlugin_Run(t *testing.T) {
	type fields struct {
		configs     []v1.Plugin
		clusterName string
	}
	type args struct {
		cluster *v2.Cluster
		phase   Phase
	}
	plugins := []v1.Plugin{
		{
			Spec: v1.PluginSpec{
				On:     "node-role.kubernetes.io/master=",
				Data:   "kubectl taint nodes node-role.kubernetes.io/master=:NoSchedule",
				Action: "PostInstall",
			},
		},
	}
	//TODO cluster is where?
	cluster := &v2.Cluster{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test",
			fields: fields{
				configs:     plugins,
				clusterName: "my-cluster",
			},
			args: args{
				cluster: cluster,
				phase:   "PostInstall",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PluginsProcessor{
				Plugins: tt.fields.configs,
				Cluster: &v2.Cluster{},
			}
			c.Cluster.Name = tt.fields.clusterName
			if err := c.Run(tt.args.cluster.GetAllIPList(), tt.args.phase); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
