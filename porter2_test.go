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

package porter2

import (
	"bufio"
	"compress/gzip"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/surgebase/glog"
)

var (
	data0 [][]rune = [][]rune{
		[]rune("abc"),
		[]rune("abcs'"),
		[]rune("abc'"),
		[]rune("abc's"),
		[]rune("abc's'"),
		[]rune("abc'sd'"),
		[]rune("abc'ss'"),
		[]rune("abc'ss''"),
	}

	expect0 [][]rune = [][]rune{
		[]rune("abc"),
		[]rune("abcs"),
		[]rune("abc"),
		[]rune("abc"),
		[]rune("abc"),
		[]rune("abc'sd"),
		[]rune("abc'ss"),
		[]rune("abc'ss'"),
	}

	data1a [][]rune = [][]rune{
		[]rune("ties"),
		[]rune("cries"),
		[]rune("ponies"),
		[]rune("caress"),
		[]rune("cats"),
		[]rune("caresses"),
	}

	expect1a [][]rune = [][]rune{
		[]rune("tie"),
		[]rune("cri"),
		[]rune("poni"),
		[]rune("caress"),
		[]rune("cat"),
		[]rune("caress"),
	}

	data1b [][]rune = [][]rune{
		[]rune("feed"),
		[]rune("agreed"),
		[]rune("plastered"),
		[]rune("bled"),
		[]rune("motoring"),
		[]rune("sing"),
		[]rune("conflated"),
		[]rune("troubled"),
		[]rune("sized"),
		[]rune("hopping"),
		[]rune("tanned"),
		[]rune("falling"),
		[]rune("hissing"),
		[]rune("fizzed"),
		[]rune("failing"),
		[]rune("filing"),
	}

	expect1b [][]rune = [][]rune{
		[]rune("feed"),
		[]rune("agree"),
		[]rune("plaster"),
		[]rune("bled"),
		[]rune("motor"),
		[]rune("sing"),
		[]rune("conflate"),
		[]rune("trouble"),
		[]rune("size"),
		[]rune("hop"),
		[]rune("tan"),
		[]rune("fall"),
		[]rune("hiss"),
		[]rune("fizz"),
		[]rune("fail"),
		[]rune("file"),
	}

	data1c [][]rune = [][]rune{
		[]rune("cry"),
		[]rune("by"),
		[]rune("say"),
		[]rune("happy"),
		[]rune("apology"),
	}

	expect1c [][]rune = [][]rune{
		[]rune("cri"),
		[]rune("by"),
		[]rune("say"),
		[]rune("happi"),
		[]rune("apologi"),
	}

	data2 [][]rune = [][]rune{
		[]rune("relational"),
		[]rune("conditional"),
		[]rune("rational"),
		[]rune("valenci"),
		[]rune("hesitanci"),
		[]rune("digitizer"),
		[]rune("conformabli"),
		[]rune("radicalli"),
		[]rune("differentli"),
		[]rune("vileli"),
		[]rune("analogousli"),
		[]rune("vietnamization"),
		[]rune("predication"),
		[]rune("operator"),
		[]rune("feudalism"),
		[]rune("decisiveness"),
		[]rune("hopefulness"),
		[]rune("callousness"),
		[]rune("formaliti"),
		[]rune("sensitiviti"),
		[]rune("sensibiliti"),
	}

	expect2 [][]rune = [][]rune{
		[]rune("relate"),
		[]rune("condition"),
		[]rune("rational"),
		[]rune("valence"),
		[]rune("hesitance"),
		[]rune("digitize"),
		[]rune("conformable"),
		[]rune("radical"),
		[]rune("different"),
		[]rune("vile"),
		[]rune("analogous"),
		[]rune("vietnamize"),
		[]rune("predicate"),
		[]rune("operate"),
		[]rune("feudal"),
		[]rune("decisive"),
		[]rune("hopeful"),
		[]rune("callous"),
		[]rune("formal"),
		[]rune("sensitive"),
		[]rune("sensible"),
	}

	data3 [][]rune = [][]rune{
		[]rune("triplicate"),
		[]rune("formative"),
		[]rune("formalize"),
		[]rune("electriciti"),
		[]rune("electrical"),
		[]rune("hopeful"),
		[]rune("goodness"),
	}

	expect3 [][]rune = [][]rune{
		[]rune("triplic"),
		[]rune("formative"),
		[]rune("formal"),
		[]rune("electric"),
		[]rune("electric"),
		[]rune("hope"),
		[]rune("good"),
	}

	data4 [][]rune = [][]rune{
		[]rune("revival"),
		[]rune("allowance"),
		[]rune("inference"),
		[]rune("airliner"),
		[]rune("gyroscopic"),
		[]rune("adjustable"),
		[]rune("defensible"),
		[]rune("irritant"),
		[]rune("replacement"),
		[]rune("adjustment"),
		[]rune("dependent"),
		[]rune("adoption"),
		[]rune("homologous"),
		[]rune("communism"),
		[]rune("activate"),
		[]rune("angulariti"),
		[]rune("homologous"),
		[]rune("effective"),
		[]rune("bowdlerize"),
	}

	expect4 [][]rune = [][]rune{
		[]rune("reviv"),
		[]rune("allow"),
		[]rune("infer"),
		[]rune("airlin"),
		[]rune("gyroscop"),
		[]rune("adjust"),
		[]rune("defens"),
		[]rune("irrit"),
		[]rune("replac"),
		[]rune("adjust"),
		[]rune("depend"),
		[]rune("adopt"),
		[]rune("homolog"),
		[]rune("communism"),
		[]rune("activ"),
		[]rune("angular"),
		[]rune("homolog"),
		[]rune("effect"),
		[]rune("bowdler"),
	}

	data5 [][]rune = [][]rune{
		[]rune("probate"),
		[]rune("rate"),
		[]rune("cease"),
		[]rune("controll"),
		[]rune("roll"),
	}

	expect5 [][]rune = [][]rune{
		[]rune("probat"),
		[]rune("rate"),
		[]rune("ceas"),
		[]rune("control"),
		[]rune("roll"),
	}

	dataRegions [][]rune = [][]rune{
		[]rune("beautiful"),
		[]rune("beauty"),
		[]rune("beau"),
		[]rune("animadversion"),
		[]rune("sprinkled"),
		[]rune("eucharist"),
	}

	expectRegions [][]int = [][]int{
		[]int{5, 7},
		[]int{5, 6},
		[]int{4, 4},
		[]int{2, 4},
		[]int{5, 9},
		[]int{3, 6},
	}

	exceptions1 map[string]string = map[string]string{
		"skis":   "ski",
		"skies":  "sky",
		"dying":  "die",
		"lying":  "lie",
		"tying":  "tie",
		"idly":   "idl",
		"gently": "gentl",
		"ugly":   "ugli",
		"early":  "earli",
		"only":   "onli",
		"singly": "singl",
		"sky":    "sky",
		"news":   "news",
		"howe":   "howe",
		"atlas":  "atlas",
		"cosmos": "cosmos",
		"bias":   "bias",
		"andes":  "andes",
	}

	exceptions2 map[string]bool = map[string]bool{
		"inning":  true,
		"outing":  true,
		"canning": true,
		"herring": true,
		"earring": true,
		"proceed": true,
		"exceed":  true,
		"succeed": true,
	}

	shortWords map[string]bool = map[string]bool{
		"rap":     true,
		"trap":    true,
		"entrap":  false,
		"bed":     true,
		"shed":    true,
		"shred":   true,
		"bead":    false,
		"embed":   false,
		"beds":    false,
		"uproot":  false,
		"bestow":  false,
		"disturb": false,
	}
)

func TestEnglishStep0(t *testing.T) {

	for i, rs := range data0 {
		//glog.Debugf("rs=%q, expected=%q", string(rs), string(expect[i]))
		assert.Equal(t, step0(rs), expect0[i])
	}
}

func TestEnglishStep1a(t *testing.T) {

	for i, rs := range data1a {
		assert.Equal(t, step1a(rs), expect1a[i])
		//glog.Debugf("rs=%q, expected=%q, got=%q", string(rs), string(expect1a[i]), string(s))
	}
}

func TestEnglishStep1b(t *testing.T) {

	for i, rs := range data1b {
		r1, _ := markR1R2(rs)
		//glog.Debugf("rs=%q, expected=%q, r1=%d", string(rs), string(expect1b[i]), r1)
		s := step1b(rs, r1)
		assert.Equal(t, s, expect1b[i])
	}
}

func TestEnglishStep1c(t *testing.T) {

	for i, rs := range data1c {
		//glog.Debugf("rs=%q, expected=%q, got=%q", string(rs), string(expect1c[i]), string(step1c(rs)))
		assert.Equal(t, step1c(rs), expect1c[i])
	}
}

func TestEnglishStep2(t *testing.T) {

	for i, rs := range data2 {
		r1, _ := markR1R2(rs)
		s := step2(rs, r1)
		//glog.Debugf("rs=%q, expected=%q, got=%q, r1=%d", string(rs), string(expect2[i]), string(s), r1)
		assert.Equal(t, s, expect2[i])
	}
}

func TestEnglishStep3(t *testing.T) {

	for i, rs := range data3 {
		r1, r2 := markR1R2(rs)
		s := step3(rs, r1, r2)
		//glog.Debugf("rs=%q, expected=%q, got=%q, r1=%d", string(rs), string(expect3[i]), string(s), r1)
		assert.Equal(t, s, expect3[i])
	}
}

func TestEnglishStep4(t *testing.T) {

	for i, rs := range data4 {
		_, r2 := markR1R2(rs)
		s := step4(rs, r2)
		//glog.Debugf("rs=%q, expected=%q, got=%q, r1=%d, r2=%d", string(rs), string(expect4[i]), string(s), r1, r2)
		assert.Equal(t, s, expect4[i])
	}
}

func TestEnglishStep5(t *testing.T) {

	for i, rs := range data5 {
		r1, r2 := markR1R2(rs)
		s := step5(rs, r1, r2)
		//glog.Debugf("rs=%q, expected=%q, got=%q, r1=%d, r2=%d", string(rs), string(expect5[i]), string(s), r1, r2)
		assert.Equal(t, s, expect5[i])
	}
}

func TestEnglishMarkR1R2(t *testing.T) {
	for i, rs := range dataRegions {
		r1, r2 := markR1R2(rs)
		//glog.Debugf("rs = %v, expected = %v, got = %v", rs, expectRegions[i], []int{r1, r2})
		assert.Equal(t, expectRegions[i], []int{r1, r2})
	}

}

func TestEnglishIsShortWord(t *testing.T) {
	for s, v := range shortWords {
		rs := []rune(s)
		r1, _ := markR1R2(rs)
		//glog.Debugf("rs=%q, r1=%d", s, r1)
		assert.Equal(t, v, isShortWord(rs, r1))
	}
}

func TestEnglishExceptions1(t *testing.T) {

	for k, v := range exceptions1 {
		rs, ex := exception1([]rune(k))
		//glog.Debugf("rs=%q, expected=%q, got=%q", k, v, string(rs))
		assert.True(t, ex)
		assert.Equal(t, []rune(v), rs)
	}

}

func TestEnglishExceptions2(t *testing.T) {

	for k, v := range exceptions2 {
		assert.Equal(t, v, exception2([]rune(k)))
	}
}

/*
func TestEnglishStem(t *testing.T) {
	fmt.Println(Stem("failure"))
}
*/

func BenchmarkEnglishStep0(b *testing.B) {

	for i := 0; i < b.N; i++ {
		for _, rs := range data0 {
			step0(rs)
		}
	}
}

func BenchmarkEnglishStep1a(b *testing.B) {

	for i := 0; i < b.N; i++ {
		for _, rs := range data1a {
			step0(rs)
		}
	}
}

func BenchmarkEnglishException1(b *testing.B) {
	var words [][]rune

	for k, _ := range exceptions1 {
		words = append(words, []rune(k))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, k := range words {
			exception1(k)
		}
	}
}

func TestEnglishVocOutput(t *testing.T) {
	inscan, infile := openFile("voc.txt")
	outscan, outfile := openFile("output.txt")
	defer infile.Close()
	defer outfile.Close()

	for inscan.Scan() {
		if !outscan.Scan() {
			break
		}

		word := inscan.Text()
		expect := outscan.Text()
		//glog.Debugf("word=%q, expect=%q", word, expect)
		actual := Stem(word)
		if actual != expect {
			glog.Debugf("word=%q, actual=%q != expect=%q", word, actual, expect)
		}
		assert.Equal(t, expect, actual)
	}
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
