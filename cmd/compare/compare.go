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

	agonopol "github.com/agonopol/go-stem"
	dchest "github.com/dchest/stemmer/porter2"
	kljensen "github.com/kljensen/snowball"
	reiver "github.com/reiver/go-porterstemmer"
	surgebase "github.com/surgebase/porter2"
)

func main() {
	actions := []string{
		"access",
		"alert",
		"allocate",
		"allow",
		"audit",
		"backup",
		"bind",
		"block",
		"cancel",
		"clean",
		"close",
		"compress",
		"connect",
		"copy",
		"create",
		"decode",
		"decompress",
		"decrypt",
		"depress",
		"detect",
		"disconnect",
		"download",
		"encode",
		"encrypt",
		"establish",
		"execute",
		"filter",
		"find",
		"free",
		"get",
		"initialize",
		"initiate",
		"install",
		"lock",
		"login",
		"logout",
		"modify",
		"move",
		"open",
		"post",
		"quarantine",
		"read",
		"release",
		"remove",
		"replicate",
		"resume",
		"save",
		"scan",
		"search",
		"start",
		"stop",
		"suspend",
		"uninstall",
		"unlock",
		"update",
		"upgrade",
		"upload",
		"violate",
		"write",
	}

	n := 10000

	now := time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			surgebase.Stem(a)
		}
	}

	since := time.Since(now)
	fmt.Println("surgebase:", since)

	//eng := porter2.Stemmer
	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			dchest.Stemmer.Stem(a)
		}
	}

	since = time.Since(now)
	fmt.Println("dchest:", since)

	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			reiver.StemString(a)
		}
	}

	since = time.Since(now)
	fmt.Println("reiver:", since)

	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			kljensen.Stem(a, "english", true)
		}
	}

	since = time.Since(now)
	fmt.Println("kljensen:", since)

	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			agonopol.Stem([]byte(a))
		}
	}

	since = time.Since(now)
	fmt.Println("agonopol:", since)
}
