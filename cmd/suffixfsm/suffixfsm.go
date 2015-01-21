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
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type node struct {
	r rune    // node value
	f bool    // final node
	s int     // state
	c []*node // children
	w string  // suffix word
}

func openFile(fname string) (*bufio.Scanner, *os.File) {
	var s *bufio.Scanner

	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasSuffix(fname, ".gz") {
		gunzip, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}

		s = bufio.NewScanner(gunzip)
	} else {
		s = bufio.NewScanner(f)
	}

	return s, f
}

func main() {
	flag.Parse()

	scan, file := openFile(flag.Arg(0))
	defer file.Close()

	s := 0 // state
	root := &node{s: s}
	nodes := append(make([]*node, 0, 10), root)

	for scan.Scan() {
		w := scan.Text()
		rs := []rune(w)
		cur := root

		for i := len(rs) - 1; i >= 0; i-- {
			r := rs[i]
			found := false
			var n *node

			for _, n = range cur.c {
				if n.r == r {
					found = true
					break
				}
			}

			if !found {
				s++
				n = &node{r: r, s: s}
				cur.c = append(cur.c, n)
				nodes = append(nodes, n)
			}

			cur = n
			if i == 0 {
				n.f = true
				n.w = w
			}
		}
	}

	fmt.Println(`var (
		l int = len(rs) // string length
		m int			// suffix length
		s int			// state
		f int			// end state of longgest suffix
		r rune			// current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {
`)

	for _, n := range nodes {
		if len(n.c) > 0 {
			fmt.Printf("case %d:\n", n.s)
			fmt.Printf("\tswitch r {\n")

			for _, c := range n.c {
				fmt.Printf("\tcase '%c':\n", c.r)
				fmt.Printf("\t\ts = %d\n", c.s)
				if c.f {
					fmt.Printf("\t\tm = %d\n", len(c.w))
					fmt.Printf("\t\tf = %d\n", c.s)
					fmt.Printf("\t\t// %s - final\n", c.w)
				}
			}

			fmt.Printf("\tdefault:\n\t\tbreak loop\n\t}\n")
		}
	}

	fmt.Printf(`default:
			break loop
		}
	}

	switch f {
`)

	for _, n := range nodes {
		if n.f {
			fmt.Printf("\tcase %d:\n", n.s)
			fmt.Printf("\t\t// %s - final\n\n", n.w)
		}
	}

	fmt.Printf("\t}\n\treturn rs\n")
}
