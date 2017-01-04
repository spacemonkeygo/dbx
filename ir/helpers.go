// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

// returns true if left is a subset of right
func fieldSetSubset(left, right []*Field) bool {
	if len(left) > len(right) {
		return false
	}
lcols:
	for _, lcol := range left {
		for _, rcol := range right {
			if lcol == rcol {
				continue lcols
			}
		}
		return false
	}
	return true
}

// returns true if left and right are equivalent (order agnostic)
func fieldSetEquivalent(left, right []*Field) bool {
	if len(left) != len(right) {
		return false
	}
	return fieldSetSubset(left, right)
}

// returns true if left is a subset of right
func fieldSetPrune(all, bad []*Field) (out []*Field) {
	for i := range all {
		if fieldSetSubset(all[i:i+1], bad) {
			continue
		}
		out = append(out, all[i])
	}
	return out
}
