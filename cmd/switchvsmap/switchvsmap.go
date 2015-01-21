// Copyright (c) 2014 Dataence, LLC. All rights reserved.
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

package main

import (
	"fmt"
	"time"
)

var amap map[rune]rune = map[rune]rune{
	'a': 'a',
	'b': 'b',
	'c': 'c',
	'd': 'd',
	'e': 'e',
	'f': 'f',
	'g': 'g',
}

func runswitch(a rune) rune {
	i := 0
	switch {
	case i == 0 && a == 'a':
		return a

	case i == 0 && a == 'b':
		return a

	case i == 0 && a == 'c':
		return a

	case i == 0 && a == 'd':
		return a

	case i == 0 && a == 'e':
		return a

	case i == 0 && a == 'f':
		return a

	case i == 0 && a == 'g':
		return a

	}

	return a
}

func runmap(a rune) rune {
	return amap[a]
}

func main() {
	now := time.Now()
	for i := 0; i < 1000000; i++ {
		runswitch('g')
	}
	fmt.Println(time.Since(now))

	now = time.Now()
	for i := 0; i < 1000000; i++ {
		runmap('g')
	}
	fmt.Println(time.Since(now))

}
