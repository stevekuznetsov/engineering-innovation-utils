package generator

import "github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"

// Student embeds api.Student and adds more state to make grouping generation easier
type Student struct {
	api.Student

	// collaborators are other students this student has already collaborated with
	collaborators map[*Student]int
}

// NewStudent creates a new student for the serializable student object
func NewStudent(student api.Student) *Student {
	return &Student{
		Student:       student,
		collaborators: map[*Student]int{},
	}
}

// ToAPIStudent converts this representation of a student to one that can be serialized
func (s *Student) ToAPIStudent() api.Student {
	return s.Student
}

// HasCollaboratedWith determines if this student has collaborated with another student.
// Students are assumed to be uniquely identifiable by their NetID.
func (s *Student) HasCollaboratedWith(student *Student) bool {
	return s.collaborators[student] > 0
}

// Collaborate marks the two students as having collaborated with each other and increments
// the counter of re-pairings if one occurred as the result of this action
func Collaborate(student, partner *Student) {
	student.collaborators[partner]++
	partner.collaborators[student]++

	if student.collaborators[partner] > 1 {
		netRepairings++
	}
}

// Uncollaborate removes all records of collaboration between the students, if any existed,
// and decrements the counter of re-pairings if one was removed as a result of this action
func Uncollaborate(student, partner *Student) {
	student.collaborators[partner]--
	if student.collaborators[partner] < 1 {
		delete(student.collaborators, partner)
	}

	partner.collaborators[student]--
	if partner.collaborators[student] < 1 {
		delete(partner.collaborators, student)
	}

	if student.collaborators[partner] > 0 {
		netRepairings--
	}
}

// Equals determines if two student objects are the same. We assume that netIDs are uniquely identifying
func (s *Student) Equals(otherStudent *Student) bool {
	return s.NetID == otherStudent.NetID
}
