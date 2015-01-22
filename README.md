porter2
=======

[![GoDoc](http://godoc.org/github.com/surgebase/porter2?status.svg)](http://godoc.org/github.com/surgebase/porter2)

Porter2 implements the [english Porter2 stemmer](http://snowball.tartarus.org/algorithms/english/stemmer.html). It is written completely using finite state machines to do suffix comparison, rather than the string-based or tree-based approaches. As a result, it is 660% faster compare to string comparison-based approach.

```
import "github.com/surgebase/porter2"

fmt.Println(porter2.Stem("seaweed")) // should get seawe
```

This implementation has been successfully validated with the dataset from http://snowball.tartarus.org/algorithms/english/

### Performance

This implementation by far has the highest performance of the various Go-based implementations, AFAICT. I tested a few of the implementations and the results are below. 

| Implementation | Time | Algorithm |
|----------------|------|-----------|
| [surgebase](https://github.com/surgebase/porter2) | 319.009358ms | Porter2 |
| [dchest](https://github.com/dchest/stemmer) | 2.106912401s | Porter2 |
| [kljensen](https://github.com/kljensen/snowball) | 5.725917198s | Porter2 |

To run the test again, you can run cmd/compare/compare.go (`go run compare.go`).

### State Machines

Most of the implementations, like the ones in the table above, rely completely on suffix string comparison. Basically there's a list of suffixes, and the code will loop through the list to see if there's a match. Given most of the time you are looking for the longest match, so you order the list so the longest is the first one. So if you are luckly, the match will be early on the list. But regardless that's a huge performance hit.

This implementation is based completely on finite state machines to perform suffix comparison. You compare each chacter of the string starting at the last character going backwards. The state machines at each step will determine what the longest suffix is. You can think of the state machine as an unrolled tree. 

However, writing large state machines can be very error-prone. So I wrote a [quick tool](https://github.com/surgebase/porter2/tree/master/cmd/suffixfsm) to generate most of the state machines. The tool basically takes a file of suffixes, creates a tree, then unrolls the tree by dumping each of the nodes. 

You can run the tool by `go run suffixfsm.go <filename>`.

### License

Copyright (c) 2014 Dataence, LLC. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
