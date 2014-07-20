package gox

import (
	"sort"
	"testing"
)

type labelledMatrix struct {
	mat   [][]bool
	names []string
	soln  []string
}

var (
	successfulCoverTests = []labelledMatrix{
		labelledMatrix{
			mat: [][]bool{
				[]bool{true, false, false, true, false, false, true},
				[]bool{true, false, false, true, false, false, false},
				[]bool{false, false, false, true, true, false, true},
				[]bool{false, false, true, false, true, true, false},
				[]bool{false, true, true, false, false, true, true},
				[]bool{false, true, false, false, false, false, true}},
			names: []string{"A", "B", "C", "D", "E", "F"},
			soln:  []string{"B", "D", "F"},
		},
		labelledMatrix{
			mat: [][]bool{
				[]bool{true, false, false, false},
				[]bool{true, true, true, false},
				[]bool{false, true, false, true},
				[]bool{false, false, true, true},
				[]bool{false, false, false, true},
			},
			names: []string{"A", "B", "C", "D", "E"},
			soln:  []string{"B", "E"},
		},
	}
)

func assertStringSliceEqual(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Fatalf("Solutions not equal length, a=%#v, b=%#v", a, b)
	}
	// sort the slices
	sort.StringSlice(a).Sort()
	sort.StringSlice(b).Sort()
	for i, val := range a {
		if val != b[i] {
			t.Fatalf("Solutions not same: a=%#v, b=%#v", a, b)
		}
	}
}

func TestExactCoverSuccess(t *testing.T) {
	for i, m := range successfulCoverTests {
		t.Log("Testing puzzle", i, " expecting success")
		prob, err := NewExactCoverProblem(m.mat, m.names)
		if err != nil {
			t.Fatalf("Error creating exact cover problem: %v", err)
		}
		if solns := prob.Solve(); len(solns) == 1 {
			t.Log("puzzle claims to be solved, compare solutions")
			assertStringSliceEqual(t, m.soln, solns[0])
		} else {
			t.Fatalf("Puzzle %d not solved, parsed puzzle=\n%v", i, prob)
		}
	}
}
