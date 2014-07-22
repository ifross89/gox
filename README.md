gox
===

Package gox provides an implementation of Knuth's Algorithm X (x). X provides a method of determining the solutions to the exact cover problem. The problem is represented by a matrix of boolean values, with the aim to select a subset of rows so that true appears in each column exactly once.

The technique used is known as "dancing links" which involves representing the matrix with doubly linked lists, allowing easy insertion and deletion of nodes, facilitating the backtracking used in the algorithm.

For more information see Knuth's paper, which can be found [here](http://arxiv.org/abs/cs/0011047).

Installation
------------

`go get github.com/ifross89/gox`
