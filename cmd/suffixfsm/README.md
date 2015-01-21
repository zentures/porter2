suffixfsm
=========

suffixfsm is a finite state machine generator for the [porter2](https://github.com/surgebase/porter2). It takes a list of suffixes, creates a tree, and then generates a FSM based on the tree. 

You can run the tool by `go run suffixfsm.go <filename>`.

The output is a function skeleton for each of the suffix lists. Then you can take the output and customize it.