/* Copyright 2017 Victor Penso, Matteo Dessalvi

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package collector

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodesMetrics(t *testing.T) {
	// Read the input data from a file
	file, err := os.Open("test_data/sinfo.txt")
	if err != nil {
		t.Fatalf("Can not open test data: %v", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Can not read test data: %v", err)
	}
	nm := ParseNodesMetrics(data)
	assert.Equal(t, 10, int(nm.idle["feature_a,feature_b"]))
	assert.Equal(t, 10, int(nm.down["feature_a,feature_b"]))
	assert.Equal(t, 40, int(nm.alloc["feature_a,feature_b"]))
	assert.Equal(t, 20, int(nm.alloc["feature_a"]))
	assert.Equal(t, 10, int(nm.down["null"]))
	assert.Equal(t, 42, int(nm.other["null"]))
	assert.Equal(t, 24, int(nm.other["feature_a"]))
	assert.Equal(t, 3, int(nm.planned["feature_a"]))
	assert.Equal(t, 5, int(nm.planned["feature_b"]))
}
