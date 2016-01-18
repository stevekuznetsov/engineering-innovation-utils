package generator

import (
	"reflect"
	"testing"
)

func TestIndiciesOfExtremes(t *testing.T) {
	var testCases = []struct {
		name                  string
		numbers               []int
		expectedLargestIndex  int
		expectedSmallestIndex int
	}{
		{
			name:                  "no duplicates",
			numbers:               []int{1, 2, 3, 4, 5},
			expectedSmallestIndex: 0,
			expectedLargestIndex:  4,
		},
		{
			name:                  "all duplicates",
			numbers:               []int{1, 1, 1, 1},
			expectedSmallestIndex: 0,
			expectedLargestIndex:  0,
		},
		{
			name:                  "some duplicates",
			numbers:               []int{2, 1, 1, 3, 5, 5},
			expectedSmallestIndex: 1,
			expectedLargestIndex:  4,
		},
		{
			name:                  "negatives",
			numbers:               []int{2, 1, -1, 3, 5, 5},
			expectedSmallestIndex: 2,
			expectedLargestIndex:  4,
		},
	}

	for _, testCase := range testCases {
		actualSmallestIndex, actualLargestIndex := indiciesOfExtremes(testCase.numbers)

		if actualSmallestIndex != testCase.expectedSmallestIndex {
			t.Errorf("%s: did not find the smallest index correctly, expected %d, got %d", testCase.name, testCase.expectedSmallestIndex, actualSmallestIndex)
		}
		if actualLargestIndex != testCase.expectedLargestIndex {
			t.Errorf("%s: did not find the largest index correctly, expected %d, got %d", testCase.name, testCase.expectedLargestIndex, actualLargestIndex)
		}
	}
}

func TestLargestInequality(t *testing.T) {
	var testCases = []struct {
		name                      string
		numbers                   []int
		expectedLargestInequality int
	}{
		{
			name:                      "no duplicates",
			numbers:                   []int{1, 2, 3, 4, 5},
			expectedLargestInequality: 4,
		},
		{
			name:                      "all duplicates",
			numbers:                   []int{1, 1, 1, 1, 1},
			expectedLargestInequality: 0,
		},
		{
			name:                      "some duplicates",
			numbers:                   []int{1, 1, 2, 3, 4},
			expectedLargestInequality: 3,
		},
		{
			name:                      "negative numbers",
			numbers:                   []int{1, 1, -2, 3, 4},
			expectedLargestInequality: 6,
		},
	}

	for _, testCase := range testCases {
		if actual, expected := largestInequality(testCase.numbers), testCase.expectedLargestInequality; actual != expected {
			t.Errorf("%s: did not correctly determine largest inequality, expected %d, got %d", testCase.name, expected, actual)
		}
	}
}

func TestIndiciesOfTwoSmallest(t *testing.T) {
	var testCases = []struct {
		name                        string
		numbers                     []int
		expectedSmallestIndex       int
		expectedSecondSmallestIndex int
	}{
		{
			name:                        "no duplicates",
			numbers:                     []int{1, 2, 3, 4, 5},
			expectedSmallestIndex:       0,
			expectedSecondSmallestIndex: 1,
		},
		{
			name:                        "all duplicates",
			numbers:                     []int{1, 1, 1, 1},
			expectedSmallestIndex:       0,
			expectedSecondSmallestIndex: 1,
		},
		{
			name:                        "some duplicates",
			numbers:                     []int{2, 1, 1, 3, 5, 5},
			expectedSmallestIndex:       1,
			expectedSecondSmallestIndex: 2,
		},
		{
			name:                        "negatives",
			numbers:                     []int{2, 1, -1, 3, 5, 5},
			expectedSmallestIndex:       2,
			expectedSecondSmallestIndex: 1,
		},
	}

	for _, testCase := range testCases {
		actualSmallestIndex, secondSmallestIndex := indiciesOfTwoSmallest(testCase.numbers)

		if actualSmallestIndex != testCase.expectedSmallestIndex {
			t.Errorf("%s: did not find the smallest index correctly, expected %d, got %d", testCase.name, testCase.expectedSmallestIndex, actualSmallestIndex)
		}
		if secondSmallestIndex != testCase.expectedSecondSmallestIndex {
			t.Errorf("%s: did not find the second smallest index correctly, expected %d, got %d", testCase.name, testCase.expectedSecondSmallestIndex, secondSmallestIndex)
		}
	}
}

func TestDetermineGroupSizes(t *testing.T) {
	var testCases = []struct {
		name                string
		numStudents         int
		optimalGroupSize    int
		preferSmallerGroups bool
		expectedGroupSizes  []int
	}{
		{
			name:                "divides into optimal groups, prefer smaller",
			numStudents:         20,
			optimalGroupSize:    4,
			preferSmallerGroups: true,
			expectedGroupSizes:  []int{4, 4, 4, 4, 4},
		},
		{
			name:                "divides into optimal groups, prefer larger",
			numStudents:         20,
			optimalGroupSize:    4,
			preferSmallerGroups: false,
			expectedGroupSizes:  []int{4, 4, 4, 4, 4},
		},
		{
			name:                "doesn't divide into optimal groups, prefer smaller",
			numStudents:         21,
			optimalGroupSize:    4,
			preferSmallerGroups: true,
			expectedGroupSizes:  []int{3, 3, 3, 4, 4, 4},
		},
		{
			name:                "doesn't divide into optimal groups, prefer larger",
			numStudents:         21,
			optimalGroupSize:    4,
			preferSmallerGroups: false,
			expectedGroupSizes:  []int{4, 4, 4, 4, 5},
		},
		{
			name:                "too few for even one group, prefer smaller",
			numStudents:         3,
			optimalGroupSize:    4,
			preferSmallerGroups: true,
			expectedGroupSizes:  []int{3},
		},
		{
			name:                "too few for even one group, prefer larger",
			numStudents:         3,
			optimalGroupSize:    4,
			preferSmallerGroups: false,
			expectedGroupSizes:  []int{3},
		},
		{
			// in this case we'd like to bring in three students to make the smallest group
			// larger, but we only have two other groups to steal from, so one group remains
			// smaller than we would like
			name:                "too few for perfect rearrangement, prefer smaller",
			numStudents:         11,
			optimalGroupSize:    5,
			preferSmallerGroups: true,
			expectedGroupSizes:  []int{3, 4, 4},
		},
		{
			// in this case we would like to offload three students from the smallest group,
			// but we only have two groups to put them in, so one group gets larger than
			// we would like
			name:                "too few for perfect rearrangement, prefer larger",
			numStudents:         13,
			optimalGroupSize:    5,
			preferSmallerGroups: false,
			expectedGroupSizes:  []int{6, 7},
		},
	}

	for _, testCase := range testCases {
		if actual, expected := determineGroupSizes(testCase.numStudents, testCase.optimalGroupSize, testCase.preferSmallerGroups), testCase.expectedGroupSizes; !reflect.DeepEqual(actual, expected) {
			t.Errorf("%s: did not create group sizes correctly, expected %v, got %v", testCase.name, expected, actual)
		}
	}
}
