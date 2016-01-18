package parser

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

func TestParseStudent(t *testing.T) {
	var testCases = []struct {
		name            string
		record          []string
		expectedStudent api.Student
		expectedError   error
	}{
		{
			name:            "normal record",
			record:          []string{"abc123@duke.edu", "LastName, FirstName"},
			expectedStudent: api.Student{FullName: "FirstName LastName", NetID: "abc123@duke.edu"},
			expectedError:   nil,
		},
		{
			name:            "too many commas",
			record:          []string{"abc123@duke.edu", "Last, Name, FirstName"},
			expectedStudent: api.Student{},
			expectedError:   errors.New(`found malformed name "Last, Name, FirstName", expected one comma, got 2`),
		},
		{
			name:            "wrong name format",
			record:          []string{"abc123@duke.edu", "FirstName LastName"},
			expectedStudent: api.Student{},
			expectedError:   errors.New(`found malformed name "FirstName LastName", expected one comma, got 0`),
		},
		{
			name:            "unescaped name string",
			record:          []string{"abc123@duke.edu", "LastName", "FirstName"},
			expectedStudent: api.Student{},
			expectedError:   fmt.Errorf("expected all records in CSV roster file to contain two columns, record %q contained %d", []string{"abc123@duke.edu", "LastName", "FirstName"}, 3),
		},
	}

	for _, testCase := range testCases {
		actualStudent, actualError := parseStudent(testCase.record)

		if !reflect.DeepEqual(actualStudent, testCase.expectedStudent) {
			t.Errorf("%s: correct student record not created:\n\twanted:\n\t%v\n\tgot:\n\t%v", testCase.name, testCase.expectedStudent, actualStudent)
		}

		if !reflect.DeepEqual(actualError, testCase.expectedError) {
			t.Errorf("%s: correct error not created:\n\twanted:\n\t%v\n\tgot:\n\t%v", testCase.name, testCase.expectedError, actualError)
		}
	}
}
