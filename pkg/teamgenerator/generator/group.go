package generator

import (
	"fmt"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

// Group mimics api.Group and adds more state to make grouping generation easier
type Group struct {
	// members are the studentst that make up the group
	members []*Student

	// DesiredSize is the number of students the group should have once populated
	DesiredSize int
}

// NewGroup creates a new group with the given desired size
func NewGroup(desiredSize int) *Group {
	return &Group{DesiredSize: desiredSize}
}

// ToAPIGroup converts this group to a serializable representation
func (g *Group) ToAPIGroup() api.Group {
	var members []api.Student
	for _, member := range g.members {
		members = append(members, member.ToAPIStudent())
	}

	return api.Group{Members: members}
}

// Contains determines if the group contains the given student. Students are assumed to be
// uniquely identifiable by their NetID
func (g *Group) Contains(student *Student) bool {
	for _, member := range g.members {
		if member.Equals(student) {
			return true
		}
	}

	return false
}

// ContainsCollaboratorsOf determines if the group contains any members that have collaborated
// with the student in question
func (g *Group) ContainsCollaboratorsOf(student *Student) bool {
	for _, member := range g.members {
		if member.HasCollaboratedWith(student) {
			return true
		}
	}

	return false
}

// AddMember adds the member to the group and updates all members' collaborator lists
// and appropriately delegates changes to the repairing counter if any need to be made
func (g *Group) AddMember(student *Student) {
	if g.Contains(student) {
		fmt.Printf("adding member to group twice!!")
	}

	for _, currentMember := range g.members {
		Collaborate(currentMember, student)
	}

	g.members = append(g.members, student)
}

// RemoveMember removes the member to the group and updates all members' collaborator lists
// and appropriately delegates changes to the repairing counter if any need to be made
func (g *Group) RemoveMember(student *Student) {
	removeIndex := -1
	for i, currentMember := range g.members {
		Uncollaborate(currentMember, student)
		if currentMember.Equals(student) {
			removeIndex = i
		}
	}

	if removeIndex > 0 {
		g.members = append(g.members[:removeIndex], g.members[removeIndex:]...)
	}
}

// IsFull determines if the group has enough members
func (g *Group) IsFull() bool {
	return len(g.members) == g.DesiredSize
}
