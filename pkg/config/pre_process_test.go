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

package config

/*func TestNewProcessorsAndRun(t *testing.T) {
	config := &v1.Config{
		Spec: v1.ConfigSpec{
			Process: "value|toJson|toBase64|toSecret",
			Data: `
config:
  username: root
  passwd: xxx
`,
		},
	}

	type args struct {
		config *v1.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test value|toJson|toBase64|toSecret",
			args:    args{config},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewProcessorsAndRun(tt.args.config); (err != nil) != tt.wantErr || tt.args.config.Spec.Data != "config: eyJwYXNzd2QiOiJ4eHgiLCJ1c3JuYW1lIjoicm9vdCJ9\n" {
				t.Errorf("NewProcessorsAndRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}*/
