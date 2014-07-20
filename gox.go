// Package gox provides an implementation of Knuth's Algorithm X (x). X provides
// a method of determining the solutions to the exact cover problem. The
// problem is represented by a matrix of boolean values, with the aim to select
// a subset of rows so that true appears in each column exactly once.

// The technique used is known as "dancing links" which involves representing
// the matrix with doubly linked lists, allowing easy insertion and deletion
// of nodes, facilitating the backtracking used in the algorithm.

// For more information see Knuth's paper, which can be found at:
// http://arxiv.org/abs/cs/0011047
package gox

import (
	"fmt"
)

// node is fundemental in the dancing links implementation of x. It serves three
// purposes:
//  1) It represents a "true" value in the exact cover matrix, with pointers to
//     neighbouring nodes.
//  2) There are special "column header" nodes, which are linked to the first
//     and last nodes in a column. These keep track of how many nodes are in
//     the column, which is an important heuristic for choosing the next column
//     in the algorithm.
//  3) A "root node" which is the entry point into the matrix. When the root's
//     left and right pointer points to itself, the matrix is empty, meaning a
//     solution has been found
type node struct {
	// left, right, up and down are the pointers to neighbouring nodes in the
	// problem
	left, right, up, down *node
	// colHead points to the column header node. This is nil if the node is a
	// root node or a column header.
	colHead *node
	// rowHead points to the row header
	rowHead *rowHeader
	// colCount is the number of nodes that are remaining in the column. This
	// value is only relevant in the column header
	colCount int
	colIndex int
}

// rowHeader contains information about each row and an entry point to the row
// so that the nodes in a row can be enumerated even if they are not active in
// the problem
type rowHeader struct {
	name  string
	index int
	first *node
}

// exactCoverProblem encapsulates all the information needed to solve the exact
// cover problem. It is private so that it can be initialized through the
// constructor, NewExactCoverProblem
type exactCoverProblem struct {
	root             *node
	colHeaders       []*node
	rowHeaders       []*rowHeader
	numRows, numCols int
	// solutionRows contains the current attempt at a solution, rows are pushed
	// and popped from the slice as attempts are made at solving the problem
	solutionRows []*rowHeader
	// solutions contains and array of row-name slices of soltions found
	solutions [][]string
	// rowsByName is a map of the row headers, by name. This allows for rows
	// to be added to the solution after the problem has been generated. This
	// is useful for e.g. sudoku, which starts with the same matrix for all
	// the puzzles but the numbers that are given can be added to the solution.
	rowsByName map[string]*rowHeader
}

// NewExactCoverProblem creates a new exact cover problem. m is a matrix of
// bools which specifies the problem to be solved. n is the names of the rows
// in the problem and are used to identify the solutions that are found.
func NewExactCoverProblem(m [][]bool, n []string) (*exactCoverProblem, error) {
	// Perform sanity checks on the inputs

	ret := &exactCoverProblem{}

	err := ret.checkInputs(m, n)
	if err != nil {
		return nil, err
	}

	// Initialize problem fields
	ret.numRows = len(m)
	ret.numCols = len(m[0]) // Safe after verification
	ret.rowsByName = make(map[string]*rowHeader)

	// Create root, ensure the column index is invalid
	ret.root = &node{colIndex: -1}
	ret.root.right = ret.root
	ret.root.left = ret.root

	ret.allocateColHeaders()
	ret.initializeColHeaders()
	err = ret.createNodes(m, n)
	if err != nil {
		return nil, err
	}
	// Now create the nodes
	return ret, nil
}

// checkInputs makes sure the inputs given are sane
func (p *exactCoverProblem) checkInputs(m [][]bool, n []string) error {
	if len(m) <= 1 {
		return fmt.Errorf("Number of columns must exceed 1")
	}

	rowLen := len(m[0])
	for i, row := range m {
		if len(row) != rowLen {
			return fmt.Errorf("All rows must be same length: rows[0]=%d, rows[%d] = %d", rowLen, i, len(row))
		}
	}
	return nil
}

// allocateColHeaders creates the column headers and adds them to the problem's
// slice so they can be iterated over before creating the necessary links
func (p *exactCoverProblem) allocateColHeaders() {
	for i := 0; i < p.numCols; i++ {
		p.colHeaders = append(p.colHeaders, &node{colIndex: i})
	}
}

// initializeColHeaders inserts the column headers into the problem
func (p *exactCoverProblem) initializeColHeaders() {
	for _, n := range p.colHeaders {
		n.up = n
		n.down = n
		n.right = p.root
		n.left = p.root.left
		p.root.left.right = n
		p.root.left = n
	}
}

// createNodes adds the problem's nodes into the linked list matrix
func (p *exactCoverProblem) createNodes(m [][]bool, n []string) error {
	for rowIndex := range m {
		// Create the row header
		rowHead := &rowHeader{index: rowIndex, name: n[rowIndex]}
		// Check for duplicate names
		if _, ok := p.rowsByName[rowHead.name]; ok {
			return fmt.Errorf("Duplicate row name present: %s", rowHead.name)
		} else {
			p.rowsByName[rowHead.name] = rowHead
		}
		p.rowHeaders = append(p.rowHeaders, rowHead)
		var firstNode *node = nil

		for colIndex, elem := range m[rowIndex] {
			colHead := p.colHeaders[colIndex]
			if elem {
				nd := &node{
					rowHead:  rowHead,
					colHead:  colHead,
					colIndex: colIndex,
				}

				// Add node to row at the right, if this is the first node
				// in the row, ensure the pointer in the row header will be set
				if firstNode == nil {
					firstNode = nd
					firstNode.left = nd
					firstNode.right = nd
				} else {
					nd.right = firstNode
					nd.left = firstNode.left
					firstNode.left.right = nd
					firstNode.left = nd
				}

				if rowHead.first == nil {
					rowHead.first = nd
				}

				// Add node to column at the bottom
				nd.down = colHead
				nd.up = colHead.up
				colHead.up.down = nd
				colHead.up = nd

				// Increment the column count
				colHead.colCount += 1
			}
		}
	}

	return nil
}

// search embodies the main structure of the algorithm. This is a recursive,
// depth-first search of the problem domain that sysematically tries rows to
// find the solutions, backtracking when the constraints of the problem can no
// longer be satisfied.
func (p *exactCoverProblem) search() {
	// Check to see if the matrix is empty, this occurs when there are no
	// more column headers
	if p.root == p.root.right {
		// Solution found, copy the current solution's row names to the slice
		// of solutions.
		soln := make([]string, len(p.solutionRows))
		for i, r := range p.solutionRows {
			soln[i] = r.name
		}
		p.solutions = append(p.solutions, soln)
		return
	}

	// Retrieve the next column to satisfy, if there are no rows in any of the
	// columns, the problem is not solvable, so backtrack
	colHead := p.nextCol()
	if colHead.colCount == 0 {
		return
	}

	p.cover(colHead)

	// Attempt to add each row in turn to the solution
	for rowNode := colHead.down; rowNode != colHead; rowNode = rowNode.down {
		// Add to partial solution
		p.pushRowToSolution(rowNode.rowHead)

		// For each node in the row, remove the all nodes in the column as
		// the constraint has been satisfied
		for rightNode := rowNode.right; rightNode != rowNode; rightNode = rightNode.right {
			p.cover(rightNode.colHead)
		}

		// search again on the reduced matrix
		p.search()

		// remove the row from the solution as either a solution has been found
		// and copied to the solutions, or the attempt was incorrect
		p.popRowFromSolution()

		// uncover the columns that were covered when the row was added to the
		// solution
		for leftNode := rowNode.left; leftNode != rowNode; leftNode = leftNode.left {
			p.uncover(leftNode.colHead)
		}

	}

	// add back the column to the matrix
	p.uncover(colHead)
}

// cover removes a column from a solution. It removes the rows from the matrix
// for which there is a node in the column
func (p *exactCoverProblem) cover(head *node) {
	// remove the column header
	head.right.left = head.left
	head.left.right = head.right

	// for each node in each row that is in the column, remove it from the
	// matrix
	for rowNode := head.down; rowNode != head; rowNode = rowNode.down {
		for rightNode := rowNode.right; rightNode != rowNode; rightNode = rightNode.right {
			rightNode.up.down = rightNode.down
			rightNode.down.up = rightNode.up

			// Update count of nodes in the column header to reflect the removal
			// of the node
			rightNode.colHead.colCount -= 1
		}
	}
}

// uncover is the reverse of cover. It adds back the removed nodes from the
// problem, allowing backtracking
func (p *exactCoverProblem) uncover(head *node) {
	// add in all the rows that were removed for the covered column
	for rowNode := head.up; rowNode != head; rowNode = rowNode.up {
		for leftNode := rowNode.left; leftNode != rowNode; leftNode = leftNode.left {
			leftNode.up.down = leftNode
			leftNode.down.up = leftNode

			// Update column node count
			leftNode.colHead.colCount += 1
		}
	}

	// add back in the column header
	head.right.left = head
	head.left.right = head
}

// nextCol picks the next column which has the least number of nodes present.
// if there are more than one node with the same number of nodes, nextCol choses
// the first it encounters when moving right from the node
func (p *exactCoverProblem) nextCol() *node {
	ret := p.root.right
	for n := ret; n != p.root; n = n.right {
		if n.colCount < ret.colCount {
			ret = n
		}
	}
	return ret
}

// pushRowToSolution adds a row to the working solution
func (p *exactCoverProblem) pushRowToSolution(r *rowHeader) {
	p.solutionRows = append(p.solutionRows, r)
}

// popRowFromSolution removes the last added row to the working solution
func (p *exactCoverProblem) popRowFromSolution() (ret *rowHeader) {
	ret, p.solutionRows = p.solutionRows[len(p.solutionRows)-1], p.solutionRows[:len(p.solutionRows)-1]
	return ret
}

// Solve starts the computation of the exact cover problem, it returns an slice
// of the solutions. The solutions are a slice of row names that were given when
// the exact cover problem was created.
func (p *exactCoverProblem) Solve() [][]string {
	p.search()
	return p.solutions
}

// Rows returns the names of all the rows associated with the problem
func (p *exactCoverProblem) Rows() []string {
	var ret []string
	for _, r := range p.rowHeaders {
		ret = append(ret, r.name)
	}
	return ret
}

// RowIsSolution allows the caller to specify rows as solutions to the problem
// before the computation of the solutions. This can be useful for problems
// which have a common matrix (e.g. sudoku), but with pre-selected rows given (
// the numbers given in a sudoku).
// Strictly, it should be the responsibility of the caller to provide the
// correct starting matrix, but there would be a lot of duplicated functionality
// for covering the correct rows of a puzzle
func (p *exactCoverProblem) RowIsSolution(name string) error {
	// find the row header
	header := p.rowsByName[name]
	if header == nil {
		return fmt.Errorf("No row found with name ", name)
	}

	// cover the columns which correspond to satisfied constraints for the row
	// given
	for rightNode := header.first.right; rightNode != header.first; rightNode = rightNode.right {
		p.cover(rightNode.colHead)
	}

	// Ensure that the first one is done, too
	p.cover(header.first.colHead)

	// Add the solution to the working solution.
	p.pushRowToSolution(header)
	return nil
}
