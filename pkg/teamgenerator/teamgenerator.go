package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/api"
	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/generator"
	"github.com/stevekuznetsov/engineering-innovation-utils/pkg/teamgenerator/parser"
)

var (
	// optimalGroupSize is the optimal number of members for groups
	optimalGroupSize int

	// preferSmallerGroups determines if smaller or larger than the optimal size
	// should be used when the class can't be evenly divided into groups
	preferSmallerGroups bool

	// priorGroupingFiles is a comma-delimited list of JSON files to be used to
	// initialize the grouping algorithm with prior groupings
	priorGroupingFiles string

	// rosterFile is a CSV file containing the roster of the class
	rosterFile string
)

const (
	defaultOptimalGroupSize    = 3
	defaultPreferSmallerGroups = false
)

func init() {
	flag.IntVar(&optimalGroupSize, "size", defaultOptimalGroupSize, "optimal group size")
	flag.BoolVar(&preferSmallerGroups, "smaller-groups", defaultPreferSmallerGroups, "prefer smaller groups")
	flag.StringVar(&priorGroupingFiles, "priors", "", "comma-delimited list of files containing prior groupings")
	flag.StringVar(&rosterFile, "roster", "", "CSV file containing class roster")
}

func main() {
	flag.Parse()
	projectNames := flag.Args()
	if len(projectNames) < 1 {
		fmt.Fprintln(os.Stderr, "teamgenerator requires at least one project name to create groups for")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "generating teams for the following projects: %v\n", projectNames)

	roster, err := parser.NewCSVRoster().Parse(rosterFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse roster file: %v\n", err)
		os.Exit(1)
	}

	generator := generator.NewClassGrouping(optimalGroupSize, preferSmallerGroups)

	var grouping api.ClassGrouping
	if len(priorGroupingFiles) > 0 {
		var priors []api.ProjectGrouping
		for _, file := range strings.Split(priorGroupingFiles, ",") {
			prior, err := parser.NewJSONProject().Parse(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to pase prior grouping file: %v\n", err)
				os.Exit(1)
			}
			priors = append(priors, prior)
		}
		grouping = generator.GenerateWithPriors(roster, priors, projectNames)
	} else {
		grouping = generator.Generate(roster, projectNames)
	}

	if err := json.NewEncoder(os.Stdout).Encode(&grouping); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode class grouping: %v\n", err)
		os.Exit(1)
	}
}
