package main

import (
	"fmt"
	"time"

	"github.com/dchest/stemmer/porter2"
	"github.com/reiver/go-porterstemmer"
	p2 "github.com/surgebase/porter2"
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

	var es p2.EnglishStemmer

	now := time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			es.Stem(a)
		}
	}

	since := time.Since(now)
	fmt.Println("surge:", since)

	//eng := porter2.Stemmer
	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			porter2.EngStem.Stem(a)
		}
	}

	since = time.Since(now)
	fmt.Println("dchest:", since)

	now = time.Now()

	for i := 0; i < n; i++ {
		for _, a := range actions {
			porterstemmer.StemString(a)
		}
	}

	since = time.Since(now)
	fmt.Println("reiver:", since)
}
