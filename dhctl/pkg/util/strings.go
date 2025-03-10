// Copyright 2021 Flant JSC
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

package util

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"
)

func RandomStrElement(list []string) (string, int) {
	// we silent gosec linter here
	// because we do not need security random number
	// for choice random element
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) //nolint:gosec
	indx := r.Intn(len(list))

	return list[indx], indx
}

func ExcludeElementFromSlice(list []string, elem string) []string {
	indx := -1
	for i, v := range list {
		if v == elem {
			indx = i
			break
		}
	}

	if indx >= 0 {
		firstPart := list[:indx]
		// need tmp slice because
		// res := append(list[:indx], list[indx+1:]...)
		// can affect source list
		tmp := make([]string, len(firstPart))
		copy(tmp, firstPart)
		res := append(tmp, list[indx+1:]...)

		return res
	}

	return list
}

func Sha256Encode(input string) string {
	hasher := sha256.New()

	hasher.Write([]byte(input))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
