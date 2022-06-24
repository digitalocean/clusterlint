/*
Copyright 2022 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package checks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupError(t *testing.T) {
	_, err := NewCheckFilter([]string{"basic"}, []string{"basic"}, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("cannot specify both include and exclude group conditions"), err)
}

func TestCheckError(t *testing.T) {
	_, err := NewCheckFilter(nil, nil, []string{"foo"}, []string{"bar"})

	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("cannot specify both include and exclude check conditions"), err)
}
