package generator

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

const (
	// maxReshuffles determines how many times a random student will be reshuffled in an attempt to move forward
	// in fleshing out a project's groups without increasing the number of second collaborations
	maxReshuffles = 1000
)

var (
	// desiredRepairings is the maximum desired number of repairings to have been made when all projects have been
	// fleshed out. This number will increase by 1 every time the algorithm tries and fails to complete the task
	// for the given desired amount.
	desiredRepairings = 0

	// netRepairings is the number of net repairings that have been created so far
	// Collaborate and Decollaborate methods on Students change this counter
	netRepairings = 0

	// numReshuffles is the number of reshuffles that have been committed so far
	numReshuffles = 0
)

func NewClassGrouping(optimalGroupSize int, preferSmallerGroups bool) ClassGrouping {
	return &classGrouping{optimalGroupSize: optimalGroupSize, preferSmallerGroups: preferSmallerGroups}
}

type classGrouping struct {
	optimalGroupSize    int
	preferSmallerGroups bool
}

// Generate generates a class grouping from a roster
func (g *classGrouping) Generate(students []api.Student, groupingNames []string) api.ClassGrouping {
	for {
		var roster []*Student
		for _, student := range students {
			roster = append(roster, NewStudent(student))
		}

		var projects []*Project
		for _, name := range groupingNames {
			projects = append(projects, NewProject(name, roster, g.optimalGroupSize, g.preferSmallerGroups))
		}

		// we're starting a new attempt at pairing, so we reset the counters
		netRepairings = 0
		numReshuffles = 0

		for _, project := range projects {
			if err := groupStudentsForProject(project, roster); err != nil {
				// the only error that can occur in this step is the algorithm
				// reaching the reshuffle quota limit. In that case, we need to
				//  increase the number of desired repairings and try again
				desiredRepairings++
				fmt.Printf("Increased the amount of desired repairings to %d after reshuffle quota was reached\n", desiredRepairings)
				continue
			}
		}

		if netRepairings <= desiredRepairings {
			fmt.Printf("Succeeded at creating groupings with %d repairings\n", netRepairings)
			var groupings []api.ProjectGrouping
			for _, finishedProject := range projects {
				groupings = append(groupings, finishedProject.ToAPIProjectGrouping())
			}

			return api.ClassGrouping{Projects: groupings}
		}
		desiredRepairings++
		fmt.Printf("Increased the amount of desired repairings to %d after grouping succeeded with too many repairings\n", desiredRepairings)
	}
}

// GenerateWithPriors generates a class grouping from a roster, taking into account prior groupings
func (g *classGrouping) GenerateWithPriors(students []api.Student, priorGroupings []api.ProjectGrouping, groupingNames []string) api.ClassGrouping {
	for {
		var roster []*Student
		associativeRoster := map[string]*Student{}
		for _, student := range students {
			internalStudent := NewStudent(student)
			roster = append(roster, internalStudent)
			associativeRoster[student.NetID] = internalStudent
		}

		var projects []*Project
		for _, name := range groupingNames {
			projects = append(projects, NewProject(name, roster, g.optimalGroupSize, g.preferSmallerGroups))
		}

		// by creating a throwaway group for all of the groups that we're recieving as prior information,
		// we can populate the collaboration lists
		for _, prior := range priorGroupings {
			for _, group := range prior.Groups {
				throwaway := NewGroup(len(group.Members))
				for _, member := range group.Members {
					if associativeRoster[member.NetID] != nil {
						// if there's someone in our group that's not on the roster, we don't care about them
						throwaway.AddMember(associativeRoster[member.NetID])
					}
				}
			}
		}

		// we're starting a new attempt at pairing, so we reset the counters
		netRepairings = 0
		numReshuffles = 0

		for _, project := range projects {
			if err := groupStudentsForProject(project, roster); err != nil {
				// the only error that can occur in this step is the algorithm
				// reaching the reshuffle quota limit. In that case, we need to
				//  increase the number of desired repairings and try again
				desiredRepairings++
				fmt.Printf("Increased the amount of desired repairings to %d after reshuffle quota was reached\n", desiredRepairings)
				continue
			}
		}

		if netRepairings <= desiredRepairings {
			fmt.Printf("Succeeded at creating groupings with %d repairings\n", netRepairings)
			var groupings []api.ProjectGrouping
			for _, finishedProject := range projects {
				groupings = append(groupings, finishedProject.ToAPIProjectGrouping())
			}

			return api.ClassGrouping{Projects: groupings}
		}
		desiredRepairings++
		fmt.Printf("Increased the amount of desired repairings to %d after grouping succeeded with too many repairings\n", desiredRepairings)
	}
}

// groupStudentsForProject will assign groups members until all groups are fulfilled, while minimizing the number of times
// any two students collaborate with each other.
// This method will return an error if the reshuffle quota is reached.
func groupStudentsForProject(project *Project, roster []*Student) error {
	groupsToFill := &GroupQueue{}
	for _, group := range project.Groups {
		groupsToFill.Enqueue(group)
	}

	for {
		if groupsToFill.IsEmpty() {
			break
		}

		if err := addMemberToGroup(project, groupsToFill, roster); err != nil {
			return err
		}
	}

	return nil
}

// addMemberToGroup adds a member to a group using the context of the given project and returns the number of
// net repairings as a result of this action as well as the number of reshuffles used in this action
// This method will return an error if the reshuffle quota is reached.
func addMemberToGroup(project *Project, groupsToFill *GroupQueue, roster []*Student) error {
	group := groupsToFill.Dequeue()
	// we want to ensure that if we haven't filled this group with this addition, that the group ends up back on the
	// queue of groups to fill
	defer func() {
		if !group.IsFull() {
			groupsToFill.Enqueue(group)
		}
	}()

	// ungrouped, fresh students are those that are ungrouped and have not collaborated with anyone in this group yet
	ungroupedFreshStudents := []*Student{}
	for _, unassignedStudent := range project.UngroupedStudents {
		if !group.ContainsCollaboratorsOf(unassignedStudent) {
			ungroupedFreshStudents = append(ungroupedFreshStudents, unassignedStudent)
		}
	}

	// TODO: make sure we're not doubling up on group members because we currently draw eligible students
	// from the total roster

	if len(ungroupedFreshStudents) != 0 {
		// if we have ungrouped and fresh students, we can just add one to our group and move on
		studentToAdd := ungroupedFreshStudents[rand.Intn(len(ungroupedFreshStudents))]
		group.AddMember(studentToAdd)
		project.MarkStudentGrouped(studentToAdd)
		return nil
	}

	if netRepairings < desiredRepairings {
		// if we don't have any ungrouped and fresh students to add to this group but we still have some of our repairing
		// quota left, we can simply add an ungrouped but stale student to our group
		studentToAdd := project.UngroupedStudents[rand.Intn(len(project.UngroupedStudents))]
		group.AddMember(studentToAdd)
		project.MarkStudentGrouped(studentToAdd)
		return nil
	}

	// if we don't have any repairing quota to use, we're going to need to undo some previously-assigned grouping
	// randomly in the hopes of moving out of this predicament. This is called a reshuffle and we limit the number
	// of times we let this occur
	numReshuffles++
	if numReshuffles >= maxReshuffles {
		return errors.New("ran out of reshuffle quota")
	}

	// first, we check to see if there are any grouped students in the class that could possibly go in this group
	// without increasing the total number of re-pairings
	potentialStudents := []*Student{}
	for _, student := range roster {
		if !group.ContainsCollaboratorsOf(student) {
			potentialStudents = append(potentialStudents, student)
		}
	}

	if len(potentialStudents) != 0 {
		// there are members of the class that could belong to this group, but belong to other groups instead.
		// we're going to remove one of them from their current group, put them into ours
		studentToPoach := potentialStudents[rand.Intn(len(potentialStudents))]
		poachStudentIntoGroup(studentToPoach, group, project, groupsToFill)
		return nil
	}

	// if there are no members of the class that haven't collaborated with anyone in this group already,
	// we need to remove members of this group until that is the case, then attempt to put someone eligible into
	// the group
	for {
		if len(potentialStudents) > 0 {
			break
		}

		unluckyStudent := group.members[rand.Intn(len(group.members))]
		group.RemoveMember(unluckyStudent)
		project.MarkStudentUngrouped(unluckyStudent)

		for _, student := range roster {
			if !group.ContainsCollaboratorsOf(student) {
				potentialStudents = append(potentialStudents, student)
			}
		}
	}

	// we've removed enough members from the group so that someone else in the class can fit in this group
	studentToPoach := potentialStudents[rand.Intn(len(potentialStudents))]
	poachStudentIntoGroup(studentToPoach, group, project, groupsToFill)
	return nil
}

// poachStudentIntoGroup removes the student to poach from their current group and adds them to the group needing a member
func poachStudentIntoGroup(studentToPoach *Student, groupNeedingMember *Group, project *Project, groupsToFill *GroupQueue) {
	previouslyGrouped := false
	for _, unluckyGroup := range project.Groups {
		if unluckyGroup.Contains(studentToPoach) {
			previouslyGrouped = true
			if unluckyGroup.IsFull() {
				// we want to enqueue only if the group is full, as in that case we know it's not in the queue
				// if the group isn't full yet, the group is already in the queue and we don't need to add it
				groupsToFill.Enqueue(unluckyGroup)
			}
			unluckyGroup.RemoveMember(studentToPoach)
		}
	}

	if !previouslyGrouped {
		// if someone asks us to poach a student that hasn't been grouped yet, we can still "poach" them but we need to
		// do the bookkeeping ourselves to mark them as having been grouped
		project.MarkStudentGrouped(studentToPoach)
	}

	groupNeedingMember.AddMember(studentToPoach)
	if !groupNeedingMember.IsFull() {
		groupsToFill.Enqueue(groupNeedingMember)
	}
}
