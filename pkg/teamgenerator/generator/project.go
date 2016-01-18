package generator

import (
	"math"
	"sort"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

// Project mimics api.Project and adds more state to make grouping generation easier
type Project struct {
	// Groups are the groups for this project
	Groups []*Group

	// Name is the name of the project for which groups are created
	Name string

	// UngroupedStudents are the students in the class that have not yet been assigned to a group for this project
	UngroupedStudents []*Student
}

// NewProject initializes a new project grouping for the given roster
func NewProject(name string, roster []*Student, optimalGroupSize int, preferSmallerGroups bool) *Project {
	project := Project{Name: name}

	for _, student := range roster {
		project.UngroupedStudents = append(project.UngroupedStudents, student)
	}

	groupSizes := determineGroupSizes(len(roster), optimalGroupSize, preferSmallerGroups)
	for _, size := range groupSizes {
		project.Groups = append(project.Groups, NewGroup(size))
	}

	return &project
}

// ToAPIProjectGrouping converts this project to a serializable format
func (p *Project) ToAPIProjectGrouping() api.ProjectGrouping {
	var groups []api.Group
	for _, group := range p.Groups {
		groups = append(groups, group.ToAPIGroup())
	}

	return api.ProjectGrouping{Name: p.Name, Groups: groups}
}

// determineGroupSizes will determine the sizes necessary for the groups with the given preference for
// smaller groups. The list of group sizes is sorted, smallest to largest.
func determineGroupSizes(numStudents, optimalGroupSize int, preferSmallerGroups bool) []int {
	if numStudents < optimalGroupSize {
		return []int{numStudents}
	}

	numOptimalGroups := int(math.Floor(float64(numStudents / optimalGroupSize)))
	numRemainingStudents := numStudents - optimalGroupSize*numOptimalGroups

	if numRemainingStudents > 0 {
		if preferSmallerGroups {
			return determineSmallerGroupSizes(optimalGroupSize, numOptimalGroups, numRemainingStudents)
		} else {
			return determineLargerGroupSizes(optimalGroupSize, numOptimalGroups, numRemainingStudents)
		}
	}

	groupSizes := []int{}
	for i := 0; i < numOptimalGroups; i++ {
		groupSizes = append(groupSizes, optimalGroupSize)
	}
	sort.Ints(groupSizes)
	return groupSizes
}

// determineSmallerGroupSizes will determine group sizes with the following logic:
// starting with a list of group sizes, we remove one member from the largest group and add them to the
// smallest group recursively, until the biggest difference in membership is at most 1.
// The list of group sizes is sorted, smallest to largest.
func determineSmallerGroupSizes(optimalGroupSize, numOptimalGroups, numRemainingStudents int) []int {
	var groupSizes []int
	for i := 0; i < numOptimalGroups; i++ {
		groupSizes = append(groupSizes, optimalGroupSize)
	}
	groupSizes = append(groupSizes, numRemainingStudents)

	for {
		if largestInequality(groupSizes) < 2 {
			break
		}

		smallestIndex, largestIndex := indiciesOfExtremes(groupSizes)
		groupSizes[largestIndex]--
		groupSizes[smallestIndex]++
	}
	sort.Ints(groupSizes)
	return groupSizes
}

// largestInequality will return the largest inequality between any two numbers in the list
func largestInequality(numbers []int) int {
	var largestInequality int
	for i := 0; i < len(numbers); i++ {
		for j := i; j < len(numbers); j++ {
			difference := int(math.Abs(float64(numbers[i] - numbers[j])))
			if difference > largestInequality {
				largestInequality = difference
			}
		}
	}
	return largestInequality
}

// indiciesOfExtremes will return the indicies of the largest and smallest numbers in the list
// if entries in the list tie for largest or smallest, this method will return the smallest index
func indiciesOfExtremes(numbers []int) (int, int) {
	var smallestIndex, largestIndex int
	smallestNumber := math.MaxInt64
	largestNumber := math.MinInt64
	for i, number := range numbers {
		if number > largestNumber {
			largestIndex = i
			largestNumber = number
		}

		if number < smallestNumber {
			smallestIndex = i
			smallestNumber = number
		}
	}
	return smallestIndex, largestIndex
}

// determineLargerGroupSizes will determine group sizes with the following logic:
// starting with a list of group sizes, we remove one member of the smallest group and add them to the largest group
// until the biggest difference in membership is at most 1.
// The list of group sizes is sorted, smallest to largest.
func determineLargerGroupSizes(optimalGroupSize, numOptimalGroups, numRemainingStudents int) []int {
	var groupSizes []int
	for i := 0; i < numOptimalGroups; i++ {
		groupSizes = append(groupSizes, optimalGroupSize)
	}
	groupSizes = append(groupSizes, numRemainingStudents)

	for {
		if largestInequality(groupSizes) < 2 {
			break
		}

		smallestIndex, secondSmallestIndex := indiciesOfTwoSmallest(groupSizes)
		groupSizes[secondSmallestIndex]++
		groupSizes[smallestIndex]--

		// we need to remove the smallest group if it's dropped to no membership
		if groupSizes[smallestIndex] == 0 {
			groupSizes = append(groupSizes[:smallestIndex], groupSizes[smallestIndex+1:]...)
		}
	}
	sort.Ints(groupSizes)
	return groupSizes
}

// indiciesOfTwoSmallest will return the indicies of the two smallest entries in the list
// If two entries tie for smallest or second smallest, the smallest index will be returned
func indiciesOfTwoSmallest(numbers []int) (int, int) {
	smallestIndex, _ := indiciesOfExtremes(numbers)

	// in order to find the second smallest number in a sane way and deal with duplicate values,
	// we just replace the smallest number with a huge number and try again
	newNumbers := []int{}
	for i, value := range numbers {
		if i == smallestIndex {
			newNumbers = append(newNumbers, math.MaxInt64)
		} else {
			newNumbers = append(newNumbers, value)
		}
	}
	secondSmallestIndex, _ := indiciesOfExtremes(newNumbers)

	return smallestIndex, secondSmallestIndex
}

// MarkStudentGrouped removes the given student from the list of unassigned students in this project
func (p *Project) MarkStudentGrouped(student *Student) {
	removeIndex := -1
	for i, unassignedStudent := range p.UngroupedStudents {
		if unassignedStudent.Equals(student) {
			removeIndex = i
		}
	}

	if removeIndex > 0 {
		p.UngroupedStudents = append(p.UngroupedStudents[:removeIndex], p.UngroupedStudents[removeIndex:]...)
	}
}

// MarkStudentUngrouped adds the given student to the list of unassigned students in this project
func (p *Project) MarkStudentUngrouped(student *Student) {
	p.UngroupedStudents = append(p.UngroupedStudents, student)
}
