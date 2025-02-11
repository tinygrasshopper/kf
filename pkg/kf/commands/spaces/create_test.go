// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spaces

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewCreateSpaceCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
		args    []string
		setup   func(t *testing.T, fakeSpaces *fake.FakeClient)
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"object passed through": {
			args: []string{"my-ns", "--container-registry=some-registry", "--domain=domain-1", "--domain=domain-2"},
			setup: func(t *testing.T, fakeSpaces *fake.FakeClient) {
				fakeSpaces.
					EXPECT().
					Create(gomock.Any()).
					Do(func(space *v1alpha1.Space) {
						testutil.AssertEqual(t, "sets name", "my-ns", space.Name)
						testutil.AssertEqual(t, "sets container registry", "some-registry", space.Spec.BuildpackBuild.ContainerRegistry)
						testutil.AssertEqual(t, "sets domains", []v1alpha1.SpaceDomain{{Domain: "domain-1", Default: true}, {Domain: "domain-2"}}, space.Spec.Execution.Domains)
					})

				fakeSpaces.EXPECT().WaitFor(gomock.Any(), "my-ns", 1*time.Second, gomock.Any()).Return(&v1alpha1.Space{}, nil)
			},
		},
		"server failure": {
			args: []string{"my-ns"},
			setup: func(t *testing.T, fakeSpaces *fake.FakeClient) {
				fakeSpaces.
					EXPECT().
					Create(gomock.Any()).
					Return(nil, errors.New("some-server-error"))
			},
			wantErr: errors.New("some-server-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeSpaces)
			}

			buffer := &bytes.Buffer{}

			c := NewCreateSpaceCommand(&config.KfParams{Namespace: "default"}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			ctrl.Finish()
		})
	}
}
