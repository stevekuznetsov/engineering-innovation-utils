package generator

import "github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"

// ClassGrouping knows how to generate a class grouping from a roster
type ClassGrouping interface {
	// Generate generates a class grouping from a roster
	Generate(students []api.Student, groupingNames []string) (grouping api.ClassGrouping)

	// GenerateWithPriors generates a class grouping from a roster, taking into account prior groupings
	GenerateWithPriors(students []api.Student, priorGroupings []api.ProjectGrouping, groupingNames []string) (grouping api.ClassGrouping)
}
