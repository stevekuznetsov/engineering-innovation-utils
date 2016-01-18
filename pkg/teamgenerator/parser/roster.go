package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

// NewCSVRoster returns a new parser that can parse a CSV file into a list of students
func NewCSVRoster() Roster {
	return &csvRoster{}
}

type csvRoster struct{}

// Parse parses a roster from a CSV file like those created by Sakai by
// navigating to 'Gradebook->Import Grades->Download Spreadsheet Template as CSV'
// This format is as follows:
// Student ID, Student Name
// [a-z0-9]+@duke.edu,"[\w\-],( [\w\-])+"
func (r *csvRoster) Parse(inputFile string) ([]api.Student, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", inputFile, err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", inputFile, err)
	}

	roster := []api.Student{}
	for _, record := range records {
		student, err := parseStudent(record)
		if err != nil {
			return nil, err
		}

		roster = append(roster, student)
	}

	return roster, nil
}

func parseStudent(record []string) (api.Student, error) {
	if len(record) != 2 {
		return api.Student{}, fmt.Errorf("expected all records in CSV roster file to contain two columns, record %q contained %d", record, len(record))
	}

	names := strings.Split(record[1], ",")
	if len(names) != 2 {
		return api.Student{}, fmt.Errorf("found malformed name %q, expected one comma, got %d", record[1], len(names)-1)
	}

	return api.Student{
		FullName: strings.Join([]string{strings.Trim(names[1], " "), names[0]}, " "),
		NetID:    record[0],
	}, nil
}
