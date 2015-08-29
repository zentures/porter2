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

// Porter2 implements the english Porter2 stemmer. It is written completely using
// finite state machines to do suffix comparison, rather than the string-based
// or tree-based approaches. As a result, it is 660% faster compare to string-based
// implementations.
//
// http://snowball.tartarus.org/algorithms/english/stemmer.html
//
// This implementation has been successfully validated with the dataset from
// http://snowball.tartarus.org/algorithms/english/
//
// Most of the implementations rely completely on suffix string comparison.
// Basically there's a list of suffixes, and the code will loop through the
// list to see if there's a match. Given most of the time you are looking for
// the longest match, so you order the list so the longest is the first one.
// So if you are luckly, the match will be early on the list. But regardless
// that's a huge performance hit.
//
// This implementation is based completely on finite state machines to perform
// suffix comparison. You compare each chacter of the string starting at the
// last character going backwards. The state machines at each step will
// determine what the longest suffix is. You can think of the state machine
// as an unrolled tree.
//
// However, writing large state machines can be very error-prone. So I wrote
// a [quick tool](https://github.com/surgebase/porter2/tree/master/cmd/suffixfsm)
// to generate most of the state machines. The tool basically takes a file of
// suffixes, creates a tree, then unrolls the tree by dumping each of the nodes.
//
// You can run the tool by `go run suffixfsm.go <filename>`.
package porter2

import "unicode"

// Stem takes a string and returns the stemmed version based on the Porter2 algorithm.
func Stem(s string) string {
	// If the word has two letters or less, leave it as it is.
	if len(s) <= 2 {
		return s
	}

	// Convert s from string to lower case rune slice
	rs := []rune(s)
	for i, r := range rs {
		rs[i] = unicode.ToLower(r)
	}

	var ex bool

	// exception1 word list
	if rs, ex = exception1(rs); ex {
		return string(rs)
	}

	rs = preclude(rs)

	r1, r2 := markR1R2(rs)

	rs = step1a(step0(rs))

	if exception2(rs) {
		return string(rs)
	}

	return string(postlude(step5(step4(step3(step2(step1c(step1b(rs, r1)), r1), r1, r2), r2), r1, r2)))
}

// Remove initial ', if present. Then set initial y, or y after a vowel, to Y.
func preclude(rs []rune) []rune {
	if rs[0] == '\'' {
		rs = rs[1:]
	}

	if rs[0] == 'y' {
		rs[0] = 'Y'
	}

	for i, r := range rs[:len(rs)-1] {
		if isVowel(r) && rs[i+1] == 'y' {
			rs[i+1] = 'Y'
		}
	}

	return rs
}

// http://snowball.tartarus.org/texts/r1r2.html
//
// R1 is the region after the first non-vowel following a vowel, or is the null
// region at the end of the word if there is no such non-vowel.
//
// R2 is the region after the first non-vowel following a vowel in R1, or is the
// null region at the end of the word if there is no such non-vowel.
//
// If the words begins gener, commun or arsen, set R1 to be the remainder of the word.
func markR1R2(rs []rune) (int, int) {
	r1 := -1

	switch rs[0] {
	case 'g':
		if len(rs) >= 5 && rs[1] == 'e' && rs[2] == 'n' && rs[3] == 'e' && rs[4] == 'r' {
			r1 = 5
		}

	case 'c':
		if len(rs) >= 6 && rs[1] == 'o' && rs[2] == 'm' && rs[3] == 'm' && rs[4] == 'u' && rs[5] == 'n' {
			r1 = 6
		}

	case 'a':
		if len(rs) >= 5 && rs[1] == 'r' && rs[2] == 's' && rs[3] == 'e' && rs[4] == 'n' {
			r1 = 5
		}
	}

	if r1 == -1 {
		r1 = markRegion(rs)
	}

	return r1, r1 + markRegion(rs[r1:])
}

func markRegion(rs []rune) int {
	if len(rs) == 0 {
		return 0
	}

	for i, r := range rs[:len(rs)-1] {
		if isVowel(r) && !isVowel(rs[i+1]) {
			return i + 2
		}
	}
	return len(rs)
}

// Search for the longest among the suffixes, and remove if found.
// '
// 's
// 's'
func step0(rs []rune) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case '\'':
				s = 1
				m = 1
				f = 1
				// ' - final
			case 's':
				s = 2
			default:
				break loop
			}
		case 1:
			switch r {
			case 's':
				s = 4
			default:
				break loop
			}
		case 2:
			switch r {
			case '\'':
				s = 3
				m = 2
				f = 3
				// 's - final
			default:
				break loop
			}
		case 4:
			switch r {
			case '\'':
				s = 5
				m = 3
				f = 5
				// 's' - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	switch f {
	case 1, 3, 5:
		rs = rs[:l-m]
	}

	return rs
}

// Search for the longest suffix among the suffixes, and perform the action indicated.
//  sses : replace by ss
//   ied : replace by i if preceded by more than one letter, otherwise by ie (so ties -> tie, cries -> cri)
//   ies : replace by i if preceded by more than one letter, otherwise by ie (so ties -> tie, cries -> cri)
//     s : delete if the preceding word part contains a vowel not immediately before the s (so gas and this retain the s, gaps and kiwis lose it)
//    us : do nothing
//    ss : do nothing
func step1a(rs []rune) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case 's':
				s = 1
				m = 1
				f = 1
				// s - final
			case 'd':
				s = 5
			default:
				break loop
			}
		case 1:
			switch r {
			case 'e':
				s = 2
			case 'u':
				s = 9
				m = 2
				f = 9
				// us - final
			case 's':
				s = 10
				m = 2
				f = 10
				// ss - final
			default:
				break loop
			}
		case 2:
			switch r {
			case 's':
				s = 3
			case 'i':
				s = 8
				m = 3
				f = 8
				// ies - final
			default:
				break loop
			}
		case 3:
			switch r {
			case 's':
				s = 4
				m = 4
				f = 4
				// sses - final
			default:
				break loop
			}
		case 5:
			switch r {
			case 'e':
				s = 6
			default:
				break loop
			}
		case 6:
			switch r {
			case 'i':
				s = 7
				m = 3
				f = 7
				// ied - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	switch f {
	case 1:
		// s - final
		if l > 2 && hasVowel(rs[:l-2]) {
			rs = rs[:l-1]
		}

	case 4:
		// sses - final
		rs = rs[:l-2]

	case 7, 8:
		// ied - final
		// ies - final
		// if there's at least 5 runes, then replace by i, otherwise by ie
		// so ties -> tie, cries -> cri
		rs = rs[:l-m]
		if l >= 5 {
			rs = append(rs, 'i')
		} else {
			rs = append(rs, []rune("ie")...)
		}

	case 9, 10:
		// us - final
		// ss - final
		// do nothing
	}

	return rs
}

// Search for the longest suffix among the suffixes, and perform the action indicated.
// 1. ingly -> see note below
// 2. eedly -> replace by ee if in R1
// 3.  edly -> see note below
// 4.   ing -> see note below
// 5.   eed -> replace by ee if in R1
// 6.    ed -> see note below
//
// Note: delete if the preceding word part contains a vowel, and after the deletion:
//       if the word ends at, bl or iz add e (so luxuriat -> luxuriate), or
//       if the word ends with a double remove the last letter (so hopp -> hop), or
//       if the word is short, add e (so hop -> hope)
func step1b(rs []rune, r1 int) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case 'y':
				s = 1
			case 'g':
				s = 9
			case 'd':
				s = 12
			default:
				break loop
			}
		case 1:
			switch r {
			case 'l':
				s = 2
			default:
				break loop
			}
		case 2:
			switch r {
			case 'g':
				s = 3
			case 'd':
				s = 6
			default:
				break loop
			}
		case 3:
			switch r {
			case 'n':
				s = 4
			default:
				break loop
			}
		case 4:
			switch r {
			case 'i':
				s = 5
				m = 5
				f = 5
				// ingly - final
			default:
				break loop
			}
		case 6:
			switch r {
			case 'e':
				s = 7
				m = 4
				f = 7
				// edly - final
			default:
				break loop
			}
		case 7:
			switch r {
			case 'e':
				s = 8
				m = 5
				f = 8
				// eedly - final
			default:
				break loop
			}
		case 9:
			switch r {
			case 'n':
				s = 10
			default:
				break loop
			}
		case 10:
			switch r {
			case 'i':
				s = 11
				m = 3
				f = 11
				// ing - final
			default:
				break loop
			}
		case 12:
			switch r {
			case 'e':
				s = 13
				m = 2
				f = 13
				// ed - final
			default:
				break loop
			}
		case 13:
			switch r {
			case 'e':
				s = 14
				m = 3
				f = 14
				// eed - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	//glog.Debugf("rs=%q, l=%d, r1=%d, m=%d, f=%d", string(rs), l, r1, m, f)

switch1b:
	switch f {
	case 5, 7, 11, 13:
		// ingly - final
		// edly - final
		// ing - final
		// ed - final

		// delete if the preceding word part contains a vowel
		if !hasVowel(rs[:l-m]) {
			break switch1b
		}
		rs = rs[:l-m]

		if len(rs) > 2 {
			r, rr := rs[len(rs)-1], rs[len(rs)-2]

			// if the word ends at, bl or iz add e (so luxuriat -> luxuriate)
			if (rr == 'a' && r == 't') || (rr == 'b' && r == 'l') || (rr == 'i' && r == 'z') {
				rs = append(rs, 'e')
				break switch1b
			}

			// if the word ends with a double remove the last letter (so hopp -> hop)
			if r == rr {
				switch r {
				case 'b', 'd', 'f', 'g', 'm', 'n', 'p', 'r', 't':
					rs = rs[:len(rs)-1]
					break switch1b
				}
			}
		}

		// if the word is short, add e (so hop -> hope)
		if isShortWord(rs, r1) {
			rs = append(rs, 'e')
			break switch1b
		}

	case 8:
		// eedly - final
		if m >= r1 {
			rs = rs[:len(rs)-3]
		}

	case 14:
		// eed - final
		if l-r1 >= m {
			rs = rs[:len(rs)-1]
		}
	}

	return rs
}

// Replace suffix y or Y by i if preceded by a non-vowel which is not the first letter
// of the word (so cry -> cri, by -> by, say -> say)
func step1c(rs []rune) []rune {
	l := len(rs)

	if l > 2 {
		switch rs[l-1] {
		case 'y', 'Y':
			if !isVowel(rs[l-2]) {
				rs[l-1] = 'i'
			}
		}
	}

	return rs
}

// Search for the longest among the following suffixes, and, if found and in R1,
// perform the action indicated.
//
//   1.  tional -> replace by tion
//   2.    enci -> replace by ence
//   3.    anci -> replace by ance
//   4.    abli -> replace by able
//   5.   entli -> replace by ent
//   6.    izer -> replace by ize
//   7. ization -> replace by ize
//   8. ational -> replace by ate
//   9.   ation -> replace by ate
//  10.    ator -> replace by ate
//  11.   alism -> replace by al
//  12.   aliti -> replace by al
//  13.    alli -> replace by al
//  14. fulness -> replace by ful
//  15.   ousli -> replace by ous
//  16. ousness -> replace by ous
//  17. iveness -> replace by ive
//  18.   iviti -> replace by ive
//  19.  biliti -> replace by ble
//  20.     bli -> replace by ble
//  21.     ogi -> replace by og if preceded by l
//  22.   fulli -> replace by ful
//  23.  lessli -> replace by less
//  24.      li -> delete if preceded by a valid li-ending
func step2(rs []rune, r1 int) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case 's':
				s = 1
			case 'l':
				s = 14
			case 'n':
				s = 21
			case 'i':
				s = 28
			case 'm':
				s = 46
			case 'r':
				s = 61
			default:
				break loop
			}
		case 1:
			switch r {
			case 's':
				s = 2
			default:
				break loop
			}
		case 2:
			switch r {
			case 'e':
				s = 3
			default:
				break loop
			}
		case 3:
			switch r {
			case 'n':
				s = 4
			default:
				break loop
			}
		case 4:
			switch r {
			case 'l':
				s = 5
			case 's':
				s = 8
			case 'e':
				s = 11
			default:
				break loop
			}
		case 5:
			switch r {
			case 'u':
				s = 6
			default:
				break loop
			}
		case 6:
			switch r {
			case 'f':
				s = 7
				m = 7
				f = 7
				// fulness - final
			default:
				break loop
			}
		case 8:
			switch r {
			case 'u':
				s = 9
			default:
				break loop
			}
		case 9:
			switch r {
			case 'o':
				s = 10
				m = 7
				f = 10
				// ousness - final
			default:
				break loop
			}
		case 11:
			switch r {
			case 'v':
				s = 12
			default:
				break loop
			}
		case 12:
			switch r {
			case 'i':
				s = 13
				m = 7
				f = 13
				// iveness - final
			default:
				break loop
			}
		case 14:
			switch r {
			case 'a':
				s = 15
			default:
				break loop
			}
		case 15:
			switch r {
			case 'n':
				s = 16
			default:
				break loop
			}
		case 16:
			switch r {
			case 'o':
				s = 17
			default:
				break loop
			}
		case 17:
			switch r {
			case 'i':
				s = 18
			default:
				break loop
			}
		case 18:
			switch r {
			case 't':
				s = 19
				m = 6
				f = 19
				// tional - final
			default:
				break loop
			}
		case 19:
			switch r {
			case 'a':
				s = 20
				m = 7
				f = 20
				// ational - final
			default:
				break loop
			}
		case 21:
			switch r {
			case 'o':
				s = 22
			default:
				break loop
			}
		case 22:
			switch r {
			case 'i':
				s = 23
			default:
				break loop
			}
		case 23:
			switch r {
			case 't':
				s = 24
			default:
				break loop
			}
		case 24:
			switch r {
			case 'a':
				s = 25
				m = 5
				f = 25
				// ation - final
			default:
				break loop
			}
		case 25:
			switch r {
			case 'z':
				s = 26
			default:
				break loop
			}
		case 26:
			switch r {
			case 'i':
				s = 27
				m = 7
				f = 27
				// ization - final
			default:
				break loop
			}
		case 28:
			switch r {
			case 't':
				s = 29
			case 'l':
				s = 34
				m = 2
				f = 34
				// li - final
			case 'c':
				s = 55
			case 'g':
				s = 69
			default:
				break loop
			}
		case 29:
			switch r {
			case 'i':
				s = 30
			default:
				break loop
			}
		case 30:
			switch r {
			case 'l':
				s = 31
			case 'v':
				s = 44
			default:
				break loop
			}
		case 31:
			switch r {
			case 'i':
				s = 32
			case 'a':
				s = 54
				m = 5
				f = 54
				// aliti - final
			default:
				break loop
			}
		case 32:
			switch r {
			case 'b':
				s = 33
				m = 6
				f = 33
				// biliti - final
			default:
				break loop
			}
		case 34:
			switch r {
			case 's':
				s = 35
			case 'l':
				s = 39
			case 't':
				s = 51
			case 'b':
				s = 59
				m = 3
				f = 59
				// bli - final
			default:
				break loop
			}
		case 35:
			switch r {
			case 's':
				s = 36
			case 'u':
				s = 42
			default:
				break loop
			}
		case 36:
			switch r {
			case 'e':
				s = 37
			default:
				break loop
			}
		case 37:
			switch r {
			case 'l':
				s = 38
				m = 6
				f = 38
				// lessli - final
			default:
				break loop
			}
		case 39:
			switch r {
			case 'u':
				s = 40
			case 'a':
				s = 68
				m = 4
				f = 68
				// alli - final
			default:
				break loop
			}
		case 40:
			switch r {
			case 'f':
				s = 41
				m = 5
				f = 41
				// fulli - final
			default:
				break loop
			}
		case 42:
			switch r {
			case 'o':
				s = 43
				m = 5
				f = 43
				// ousli - final
			default:
				break loop
			}
		case 44:
			switch r {
			case 'i':
				s = 45
				m = 5
				f = 45
				// iviti - final
			default:
				break loop
			}
		case 46:
			switch r {
			case 's':
				s = 47
			default:
				break loop
			}
		case 47:
			switch r {
			case 'i':
				s = 48
			default:
				break loop
			}
		case 48:
			switch r {
			case 'l':
				s = 49
			default:
				break loop
			}
		case 49:
			switch r {
			case 'a':
				s = 50
				m = 5
				f = 50
				// alism - final
			default:
				break loop
			}
		case 51:
			switch r {
			case 'n':
				s = 52
			default:
				break loop
			}
		case 52:
			switch r {
			case 'e':
				s = 53
				m = 5
				f = 53
				// entli - final
			default:
				break loop
			}
		case 55:
			switch r {
			case 'n':
				s = 56
			default:
				break loop
			}
		case 56:
			switch r {
			case 'e':
				s = 57
				m = 4
				f = 57
				// enci - final
			case 'a':
				s = 58
				m = 4
				f = 58
				// anci - final
			default:
				break loop
			}
		case 59:
			switch r {
			case 'a':
				s = 60
				m = 4
				f = 60
				// abli - final
			default:
				break loop
			}
		case 61:
			switch r {
			case 'e':
				s = 62
			case 'o':
				s = 65
			default:
				break loop
			}
		case 62:
			switch r {
			case 'z':
				s = 63
			default:
				break loop
			}
		case 63:
			switch r {
			case 'i':
				s = 64
				m = 4
				f = 64
				// izer - final
			default:
				break loop
			}
		case 65:
			switch r {
			case 't':
				s = 66
			default:
				break loop
			}
		case 66:
			switch r {
			case 'a':
				s = 67
				m = 4
				f = 67
				// ator - final
			default:
				break loop
			}
		case 69:
			switch r {
			case 'o':
				s = 70
				m = 3
				f = 70
				// ogi - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	if l-r1 < m {
		return rs
	}

	switch f {
	case 7, 10, 13:
		// fulness - final
		// ousness - final
		// iveness - final
		rs = rs[:l-4]

	case 19, 38, 41, 43, 53, 68:
		// tional - final
		// lessli - final
		// fulli - final
		// ousli - final
		// entli - final
		// alli - final
		rs = rs[:l-2]

	case 20, 27:
		// ational - final
		// ization - final
		rs[l-5] = 'e'
		rs = rs[:l-4]

	case 25, 45:
		// ation - final
		// iviti - final
		rs[l-3] = 'e'
		rs = rs[:l-2]

	case 33:
		// biliti - final
		rs[l-5] = 'l'
		rs[l-4] = 'e'
		rs = rs[:l-3]

	case 34:
		// li - final
		if l > 2 {
			switch rs[l-3] {
			case 'c', 'd', 'e', 'g', 'h', 'k', 'm', 'n', 'r', 't':
				rs = rs[:l-2]
			}
		}

	case 50, 54:
		// alism - final
		// aliti - final
		rs = rs[:l-3]

	case 57, 58, 59, 60:
		// enci - final
		// anci - final
		// abli - final
		// bli - final
		rs[l-1] = 'e'

	case 64:
		// izer - final
		rs = rs[:l-1]

	case 67:
		// ator - final
		rs[l-2] = 'e'
		rs = rs[:l-1]

	case 70:
		// ogi - final
		if l > 3 && rs[l-4] == 'l' {
			rs = rs[:l-1]
		}

	}

	return rs
}

// Search for the longest among the following suffixes, and, if found and in R1,
// perform the action indicated.
//
// 1.  tional -> replace by tion
// 2. ational -> replace by ate
// 3.   alize -> replace by al
// 4.   icate -> replace by ic
// 5.   iciti -> replace by ic
// 6.    ical -> replace by ic
// 7.     ful -> delete
// 8.    ness -> delete
// 9.   ative -> delete if in R2
func step3(rs []rune, r1, r2 int) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case 'l':
				s = 1
			case 'e':
				s = 8
			case 'i':
				s = 17
			case 's':
				s = 26
			default:
				break loop
			}
		case 1:
			switch r {
			case 'a':
				s = 2
			case 'u':
				s = 24
			default:
				break loop
			}
		case 2:
			switch r {
			case 'n':
				s = 3
			case 'c':
				s = 22
			default:
				break loop
			}
		case 3:
			switch r {
			case 'o':
				s = 4
			default:
				break loop
			}
		case 4:
			switch r {
			case 'i':
				s = 5
			default:
				break loop
			}
		case 5:
			switch r {
			case 't':
				s = 6
				m = 6
				f = 6
				// tional - final
			default:
				break loop
			}
		case 6:
			switch r {
			case 'a':
				s = 7
				m = 7
				f = 7
				// ational - final
			default:
				break loop
			}
		case 8:
			switch r {
			case 'z':
				s = 9
			case 't':
				s = 13
			case 'v':
				s = 30
			default:
				break loop
			}
		case 9:
			switch r {
			case 'i':
				s = 10
			default:
				break loop
			}
		case 10:
			switch r {
			case 'l':
				s = 11
			default:
				break loop
			}
		case 11:
			switch r {
			case 'a':
				s = 12
				m = 5
				f = 12
				// alize - final
			default:
				break loop
			}
		case 13:
			switch r {
			case 'a':
				s = 14
			default:
				break loop
			}
		case 14:
			switch r {
			case 'c':
				s = 15
			default:
				break loop
			}
		case 15:
			switch r {
			case 'i':
				s = 16
				m = 5
				f = 16
				// icate - final
			default:
				break loop
			}
		case 17:
			switch r {
			case 't':
				s = 18
			default:
				break loop
			}
		case 18:
			switch r {
			case 'i':
				s = 19
			default:
				break loop
			}
		case 19:
			switch r {
			case 'c':
				s = 20
			default:
				break loop
			}
		case 20:
			switch r {
			case 'i':
				s = 21
				m = 5
				f = 21
				// iciti - final
			default:
				break loop
			}
		case 22:
			switch r {
			case 'i':
				s = 23
				m = 4
				f = 23
				// ical - final
			default:
				break loop
			}
		case 24:
			switch r {
			case 'f':
				s = 25
				m = 3
				f = 25
				// ful - final
			default:
				break loop
			}
		case 26:
			switch r {
			case 's':
				s = 27
			default:
				break loop
			}
		case 27:
			switch r {
			case 'e':
				s = 28
			default:
				break loop
			}
		case 28:
			switch r {
			case 'n':
				s = 29
				m = 4
				f = 29
				// ness - final
			default:
				break loop
			}
		case 30:
			switch r {
			case 'i':
				s = 31
			default:
				break loop
			}
		case 31:
			switch r {
			case 't':
				s = 32
			default:
				break loop
			}
		case 32:
			switch r {
			case 'a':
				s = 33
				m = 5
				f = 33
				// ative - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	//glog.Debugf("rs=%q, l=%d, r1=%d, r2=%d, m=%d, s=%d, f=%d", string(rs), l, r1, r2, m, s, f)

	// if not found and in R1, do nothing
	if l-r1 < m {
		return rs
	}

	switch f {
	case 6, 23:
		// tional - final
		// ical - final
		rs = rs[:l-2]

	case 7:
		// ational - final
		rs[l-5] = 'e'
		rs = rs[:l-4]

	case 12, 16, 21:
		// alize - final
		// icate - final
		// iciti - final
		rs = rs[:l-3]

	case 25, 29:
		// ful - final
		// ness - final
		rs = rs[:l-m]

	case 33:
		// ative - final
		// delete if in R2
		if l-r2 >= m {
			rs = rs[:l-m]
		}
	}

	return rs
}

// Search for the longest among the following suffixes, and, if found and in R2,
// perform the action indicated.
//
//   1.  able -> delete
//   2.    al -> delete
//   3.  ance -> delete
//   4.   ant -> delete
//   5.   ate -> delete
//   6. ement -> delete
//   7.  ence -> delete
//   8.   ent -> delete
//   9.    er -> delete
//  10.  ible -> delete
//  11.    ic -> delete
//  12.   ism -> delete
//  13.   iti -> delete
//  14.   ive -> delete
//  15.   ize -> delete
//  16.  ment -> delete
//  17.   ous -> delete
//  18.   ion -> delete if preceded by s or t
func step4(rs []rune, r2 int) []rune {
	var (
		l int  = len(rs) // string length
		m int            // suffix length
		s int            // state
		f int            // end state of longgest suffix
		r rune           // current rune
	)

loop:
	for i := 0; i < l; i++ {
		r = rs[l-i-1]

		switch s {

		case 0:
			switch r {
			case 'e':
				s = 1
			case 'l':
				s = 5
			case 't':
				s = 10
			case 'r':
				s = 19
			case 'c':
				s = 22
			case 'm':
				s = 24
			case 'i':
				s = 27
			case 's':
				s = 34
			case 'n':
				s = 37
			default:
				break loop
			}
		case 1:
			switch r {
			case 'l':
				s = 2
			case 'c':
				s = 7
			case 't':
				s = 13
			case 'v':
				s = 30
			case 'z':
				s = 32
			default:
				break loop
			}
		case 2:
			switch r {
			case 'b':
				s = 3
			default:
				break loop
			}
		case 3:
			switch r {
			case 'a':
				s = 4
				m = 4
				f = 4
				// able - final
			case 'i':
				s = 21
				m = 4
				f = 21
				// ible - final
			default:
				break loop
			}
		case 5:
			switch r {
			case 'a':
				s = 6
				m = 2
				f = 6
				// al - final
			default:
				break loop
			}
		case 7:
			switch r {
			case 'n':
				s = 8
			default:
				break loop
			}
		case 8:
			switch r {
			case 'a':
				s = 9
				m = 4
				f = 9
				// ance - final
			case 'e':
				s = 18
				m = 4
				f = 18
				// ence - final
			default:
				break loop
			}
		case 10:
			switch r {
			case 'n':
				s = 11
			default:
				break loop
			}
		case 11:
			switch r {
			case 'a':
				s = 12
				m = 3
				f = 12
				// ant - final
			case 'e':
				s = 15
				m = 3
				f = 15
				// ent - final
			default:
				break loop
			}
		case 13:
			switch r {
			case 'a':
				s = 14
				m = 3
				f = 14
				// ate - final
			default:
				break loop
			}
		case 15:
			switch r {
			case 'm':
				s = 16
				m = 4
				f = 16
				// ment - final
			default:
				break loop
			}
		case 16:
			switch r {
			case 'e':
				s = 17
				m = 5
				f = 17
				// ement - final
			default:
				break loop
			}
		case 19:
			switch r {
			case 'e':
				s = 20
				m = 2
				f = 20
				// er - final
			default:
				break loop
			}
		case 22:
			switch r {
			case 'i':
				s = 23
				m = 2
				f = 23
				// ic - final
			default:
				break loop
			}
		case 24:
			switch r {
			case 's':
				s = 25
			default:
				break loop
			}
		case 25:
			switch r {
			case 'i':
				s = 26
				m = 3
				f = 26
				// ism - final
			default:
				break loop
			}
		case 27:
			switch r {
			case 't':
				s = 28
			default:
				break loop
			}
		case 28:
			switch r {
			case 'i':
				s = 29
				m = 3
				f = 29
				// iti - final
			default:
				break loop
			}
		case 30:
			switch r {
			case 'i':
				s = 31
				m = 3
				f = 31
				// ive - final
			default:
				break loop
			}
		case 32:
			switch r {
			case 'i':
				s = 33
				m = 3
				f = 33
				// ize - final
			default:
				break loop
			}
		case 34:
			switch r {
			case 'u':
				s = 35
			default:
				break loop
			}
		case 35:
			switch r {
			case 'o':
				s = 36
				m = 3
				f = 36
				// ous - final
			default:
				break loop
			}
		case 37:
			switch r {
			case 'o':
				s = 38
			default:
				break loop
			}
		case 38:
			switch r {
			case 'i':
				s = 39
				m = 3
				f = 39
				// ion - final
			default:
				break loop
			}
		default:
			break loop
		}
	}

	//glog.Debugf("rs=%q, l=%d, r2=%d, m=%d, f=%d", string(rs), l, r2, m, f)
	if l-r2 < m {
		return rs
	}

	switch f {
	case 4, 6, 9, 12, 14, 15, 16, 17, 18, 20, 21, 23, 26, 29, 31, 33, 36:
		// able - final
		// al - final
		// ance - final
		// ant - final
		// ate - final
		// ent - final
		// ment - final
		// ement - final
		// ence - final
		// er - final
		// ible - final
		// ic - final
		// ism - final
		// iti - final
		// ive - final
		// ize - final
		// ous - final
		rs = rs[:l-m]

	case 39:
		// ion - final
		if l >= 4 && (rs[l-4] == 's' || rs[l-4] == 't') {
			rs = rs[:l-3]
		}

	}

	return rs
}

// Search for the the following suffixes, and, if found, perform the action indicated.
//
// e -> delete if in R2, or in R1 and not preceded by a short syllable
// l -> delete if in R2 and preceded by l
func step5(rs []rune, r1, r2 int) []rune {
	l := len(rs)
	if l < 1 {
		return rs
	}

	r := rs[l-1]
	switch r {
	case 'e':
		// in R2, delete
		if l-r2 > 0 {
			return rs[:l-1]
		}

		// not in R1, quit
		if l-r1 < 1 {
			return rs
		}

		// in R1, test to see if preceded by a short syllable
		if !isShortSyllable(rs[:l-1]) {
			return rs[:l-1]
		}

	case 'l':
		if l > 1 && l-r2 > 0 && rs[l-2] == 'l' {
			return rs[:l-1]
		}
	}

	return rs
}

// Finally, turn any remaining Y letters in the word back into lower case.
func postlude(rs []rune) []rune {
	for i, r := range rs {
		if r == 'Y' {
			rs[i] = 'y'
		}
	}

	return rs
}

// word exceptions list 1. Can't do a map since we have a []rune, and []rune cannot
// be a key to the map..argh..
//
// Returns true if word is an exception, false if not. The replacement word is
// returned if true. Otherwise the same word is returned if false.
//
//  andes -> andes
//  atlas -> atlas
//  bias -> bias
//  cosmos -> cosmos
//  dying -> die
//  early -> earli
//  gently -> gentl
//  howe -> howe
//  idly -> idl
//  lying -> lie
//  news -> news
//  only -> onli
//  singly -> singl
//  skies -> sky
//  skis -> ski
//  sky -> sky
//  tying -> tie
//  ugly -> ugli
func exception1(rs []rune) ([]rune, bool) {
	l := len(rs)
	if l > 6 {
		return rs, false
	}

	switch rs[l-1] {
	case 's', 'g', 'y', 'e':

	default:
		return rs, false
	}

	switch rs[0] {
	case 'a':
		if l != 5 {
			return rs, false
		}

		if rs[1] == 'n' && rs[2] == 'd' && rs[3] == 'e' && rs[4] == 's' {
			return rs, true
		} else if rs[1] == 't' && rs[2] == 'l' && rs[3] == 'a' && rs[4] == 's' {
			return rs, true
		}

		return rs, false

	case 'b':
		if l == 4 && rs[1] == 'i' && rs[2] == 'a' && rs[3] == 's' {
			return rs, true
		}

		return rs, false

	case 'c':
		if l == 6 && rs[1] == 'o' && rs[2] == 's' && rs[3] == 'm' && rs[4] == 'o' && rs[5] == 's' {
			return rs, true
		}

		return rs, false

	case 'd', 'l', 't':
		if l == 5 && rs[1] == 'y' && rs[2] == 'i' && rs[3] == 'n' && rs[4] == 'g' {
			rs[1], rs[2] = 'i', 'e'
			return rs[:3], true
		}

		return rs, false

	case 'e':
		if l == 5 && rs[1] == 'a' && rs[2] == 'r' && rs[3] == 'l' && rs[4] == 'y' {
			rs[4] = 'i'
			return rs, true
		}

		return rs, false

	case 'g':
		if l == 6 && rs[1] == 'e' && rs[2] == 'n' && rs[3] == 't' && rs[4] == 'l' && rs[5] == 'y' {
			return rs[:5], true
		}

		return rs, false

	case 'h':
		if l == 4 && rs[1] == 'o' && rs[2] == 'w' && rs[3] == 'e' {
			return rs, true
		}

		return rs, false

	case 'i':
		if l == 4 && rs[1] == 'd' && rs[2] == 'l' && rs[3] == 'y' {
			return rs[:3], true
		}

		return rs, false

	case 'n':
		if l == 4 && rs[1] == 'e' && rs[2] == 'w' && rs[3] == 's' {
			return rs, true
		}

		return rs, false

	case 'o':
		if l == 4 && rs[1] == 'n' && rs[2] == 'l' && rs[3] == 'y' {
			rs[3] = 'i'
			return rs, true
		}

		return rs, false

	case 's':
		switch rs[1] {
		case 'i':
			if l == 6 && rs[2] == 'n' && rs[3] == 'g' && rs[4] == 'l' && rs[5] == 'y' {
				return rs[:5], true
			}

			return rs, false

		case 'k':
			if l == 3 && rs[2] == 'y' {
				return rs, true
			} else if l == 4 && rs[2] == 'i' && rs[3] == 's' {
				rs = rs[:3]
				return rs, true
			} else if l == 5 && rs[2] == 'i' && rs[3] == 'e' && rs[4] == 's' {
				rs[2] = 'y'
				return rs[:3], true
			}

		default:
			return rs, false
		}

	case 'u':
		if l == 4 && rs[1] == 'g' && rs[2] == 'l' && rs[3] == 'y' {
			rs[3] = 'i'
			return rs, true
		}

		return rs, false
	}

	return rs, false
}

// Following step 1a, leave the following invariant,
//
//  inning
//  outing
// canning
// herring
// earring
// proceed
//  exceed
// succeed
func exception2(rs []rune) bool {
	l := len(rs)
	if l != 6 && l != 7 {
		return false
	}

	switch rs[l-1] {
	case 'g', 'd':

	default:
		return false
	}

	switch rs[0] {
	case 'i':
		// inning
		if l != 6 || rs[1] != 'n' || rs[2] != 'n' || rs[3] != 'i' || rs[4] != 'n' || rs[5] != 'g' {
			return false
		}

	case 'o':
		// outing
		if l != 6 || rs[1] != 'u' || rs[2] != 't' || rs[3] != 'i' || rs[4] != 'n' || rs[5] != 'g' {
			return false
		}

	case 'c':
		// canning
		if l != 7 || rs[1] != 'a' || rs[2] != 'n' || rs[3] != 'n' || rs[4] != 'i' || rs[5] != 'n' || rs[6] != 'g' {
			return false
		}

	case 'h':
		// herring
		if l != 7 || rs[1] != 'e' || rs[2] != 'r' || rs[3] != 'r' || rs[4] != 'i' || rs[5] != 'n' || rs[6] != 'g' {
			return false
		}

	case 'e':
		switch l {
		case 7:
			// earring
			if rs[1] != 'a' || rs[2] != 'r' || rs[3] != 'r' || rs[4] != 'i' || rs[5] != 'n' || rs[6] != 'g' {
				return false
			}

		case 6:
			// exceed
			if rs[1] != 'x' || rs[2] != 'c' || rs[3] != 'e' || rs[4] != 'e' || rs[5] != 'd' {
				return false
			}
		}

	case 'p':
		// proceed
		if l != 7 || rs[1] != 'r' || rs[2] != 'o' || rs[3] != 'c' || rs[4] != 'e' || rs[5] != 'e' || rs[6] != 'd' {
			return false
		}

	case 's':
		// succeed
		if l != 7 || rs[1] != 'u' || rs[2] != 'c' || rs[3] != 'c' || rs[4] != 'e' || rs[5] != 'e' || rs[6] != 'd' {
			return false
		}

	default:
		return false
	}

	return true
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		return true
	}

	return false
}

func hasVowel(rs []rune) bool {
	for _, r := range rs {
		if isVowel(r) {
			return true
		}
	}

	return false
}

// A word is called short if it ends in a short syllable, and if R1 is null.
//
// Define a short syllable in a word as either
//  (a) a vowel followed by a non-vowel other than w, x or Y and preceded by a non-vowel, or
//  (b) a vowel at the beginning of the word followed by a non-vowel.
func isShortWord(rs []rune, r1 int) bool {
	if r1 < len(rs) {
		return false
	}

	return isShortSyllable(rs)
}

func isShortSyllable(rs []rune) bool {
	l := len(rs)

	switch l {
	case 0, 1:
		// do nothing

	case 2:
		//  (b) a vowel at the beginning of the word followed by a non-vowel.
		if isVowel(rs[0]) && !isVowel(rs[1]) {
			return true
		}

	default:
		r, rr, rrr := rs[len(rs)-1], rs[len(rs)-2], rs[len(rs)-3]

		//  (a) a vowel followed by a non-vowel other than w, x or Y and preceded by a non-vowel, or
		// N v N
		if !isVowel(rrr) && isVowel(rr) && (!isVowel(r) && r != 'w' && r != 'x' && r != 'Y') {
			return true
		}
	}

	return false
}
