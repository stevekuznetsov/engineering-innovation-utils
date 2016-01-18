package parser

import "github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"

// Roster knows how to parse a roster of students from a file
type Roster interface {
	// Parse parses a roster of students from a file
	Parse(inputFile string) (roster []api.Student, err error)
}

// Project knows how to parse a project grouping from a file
type Project interface {
	// Parse parses a project grouping from a file
	Parse(inputFile string) (project api.ProjectGrouping, err error)
}
