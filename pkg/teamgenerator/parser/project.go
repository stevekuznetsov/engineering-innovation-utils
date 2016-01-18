package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
)

func NewJSONProject() Project {
	return &jsonProject{}
}

type jsonProject struct{}

// Parse uses Go's JSON decoding to decode the contents of the intput file into the API project object
func (p *jsonProject) Parse(inputFile string) (api.ProjectGrouping, error) {
	var project api.ProjectGrouping

	file, err := os.Open(inputFile)
	if err != nil {
		return project, fmt.Errorf("failed to open %q: %v", inputFile, err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&project)
	if err != nil {
		return project, fmt.Errorf("failed to decode JSON from %q: %v", inputFile, err)
	}

	return project, nil
}
