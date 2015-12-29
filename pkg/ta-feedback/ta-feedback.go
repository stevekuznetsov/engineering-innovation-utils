package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"
)

var (
	// outputDirectory is the directory into which output files will be placed
	outputDirectory string
)

const (
	// defaultOutputDirectory defaults the output directory to the working directory
	defaultOutputDirectory = "."
)

func init() {
	flag.StringVar(&outputDirectory, "o", defaultOutputDirectory, "where to put output files")
}

// main parses the CSV file, identifying the TAs being rated and collating student responses for them,
// creates a TeX file reporting the outcome and runs LaTeX to generate a PDF report for each TA
func main() {
	flag.Parse()
	arguments := flag.Args()
	if len(arguments) != 1 {
		fmt.Fprintln(os.Stderr, "ta-feedback requires one argument (the CSV to parse)")
		os.Exit(1)
	}

	csvFile, err := os.Open(arguments[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading the CSV file: %v\n", err)
		os.Exit(1)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	headers, err := reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing CSV headers from file: %v\n", err)
		os.Exit(1)
	}

	if (len(headers)-1)%9 != 0 {
		// Each TA feedback section takes up nine columns. The first column is a timestamp we ignore.
		// If the salient data (i.e. not the timestamp) isn't divisible by nine, we have an issue.
		fmt.Fprintln(os.Stderr, "CSV file did not contain the correct amount of columns to divide into a whole number of TAs")
		os.Exit(1)
	}

	var organizedResponses []*TAFeedback

	for i := 1; i <= len(headers)-9; i += 9 {
		record, err := initializeRecord(headers[i : i+9])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing record for TA: %v\n", err)
			os.Exit(1)
		}
		organizedResponses = append(organizedResponses, record)
	}

	responseData, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response data from CSV file: %v\n", err)
	}

	for _, response := range responseData {
		i := 1
		for _, record := range organizedResponses {
			if err := incorporateData(record, response[i:i+9]); err != nil {
				fmt.Fprintf(os.Stderr, "Error incorporating data into record: %v\n", err)
				os.Exit(1)
			}
			i += 9
		}
	}

	for _, record := range organizedResponses {
		texFile, err := generateTeXFile(record, outputDirectory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating report file for %q: %v\n", record.Name, err)
			os.Exit(1)
		}

		output, err := exec.Command("latex", "--output-directory="+outputDirectory, "--output-format=pdf", texFile).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running LaTeX: %v\nCombined output:%s\n", err, string(output))
		} else {
			fmt.Fprintf(os.Stdout, "Created report for TA %q at %q\n", record.Name, texFile[:len(texFile)-4]+".pdf")
		}
	}
}

// TAFeedback holds all of the feedback given for any one TA
type TAFeedback struct {
	// Name is the name of the TA
	Name string

	// RatingOrder is the order in which ratings are listed for the TA
	RatingOrder []string

	// Ratings is a map of rating criterion to a list of ratings given
	Ratings map[string][]int

	// PositiveFeedback is a list of feedback lines about what the TA did well
	PositiveFeedback []string

	// PositiveFeedback is a list of feedback lines about what the TA could improve on
	NegativeFeedback []string

	// PositiveFeedback is an optional list of feedback lines for other comments
	OtherFeedback []string
}

var (
	// this regex will match the name of the TA as placed in the CSV column headers
	identifyingRegex = regexp.MustCompile(`think ([\w ]+) has`)

	// this regex will determine the attribute being rated from a rating CSV column header
	ratingRegex = regexp.MustCompile(`: \[([\w ]+)\]`)
)

// initializeRecord initalizes a TAFeedback record from a set of CSV column headers
func initializeRecord(headers []string) (*TAFeedback, error) {
	record := &TAFeedback{Ratings: map[string][]int{}}
	if len(headers) != 9 {
		return record, fmt.Errorf("expected 9 header entries, for %d", len(headers))
	}

	identifyingHeader := headers[0]
	if matches := identifyingRegex.FindStringSubmatch(identifyingHeader); len(matches) > 1 && len(matches[1]) > 0 {
		name := matches[1]
		record.Name = name
	} else {
		return record, fmt.Errorf("CSV header did not contain a TA name: %q", identifyingHeader)
	}

	ratingHeaders := headers[0:6]
	for _, ratingHeader := range ratingHeaders {
		if matches := ratingRegex.FindStringSubmatch(ratingHeader); len(matches) > 1 && len(matches[1]) > 0 {
			criterion := matches[1]
			record.RatingOrder = append(record.RatingOrder, criterion)
			record.Ratings[criterion] = []int{0, 0, 0, 0, 0} // we need to initialize rating counts to 0
		} else {
			return record, fmt.Errorf("CSV header did not contain a rating criterion: %q", ratingHeader)
		}
	}
	return record, nil
}

// incorporateData incorporates CSV response data into a TAFeedback record
func incorporateData(record *TAFeedback, data []string) error {
	if len(data) != 9 {
		return fmt.Errorf("expected 9 data entries, got %d", len(data))
	}

	i := 0
	for _, criterion := range record.RatingOrder {
		rating, err := parseRating(data[i])
		if err != nil {
			return fmt.Errorf("error parsing rating response: %v", err)
		}
		record.Ratings[criterion][rating] += 1
		i++
	}

	// Often students are lazy and give bogus responses to this section,
	// using whitespace and/or single-character responses. We don't want
	// that in the report.
	if len(strings.TrimSpace(data[6])) > 2 {
		record.PositiveFeedback = append(record.PositiveFeedback, data[6])
	}

	if len(strings.TrimSpace(data[7])) > 2 {
		record.NegativeFeedback = append(record.NegativeFeedback, data[7])
	}

	if len(strings.TrimSpace(data[8])) > 2 {
		record.OtherFeedback = append(record.OtherFeedback, data[8])
	}

	return nil
}

// The following constants act as an enum for rating responses
const (
	VeryPoor = iota
	Poor
	Adequate
	Good
	VeryGood
)

func parseRating(rating string) (int, error) {
	switch rating {
	case "Very Poor":
		return VeryPoor, nil
	case "Poor":
		return Poor, nil
	case "Adequate":
		return Adequate, nil
	case "Good":
		return Good, nil
	case "Very Good":
		return VeryGood, nil
	default:
		return -1, fmt.Errorf("rating not recognized: %q", rating)
	}
}

const (
	reportOutline = `\documentclass{article}
\usepackage{microtype,pgfplots}
\pgfplotstableset{col sep=comma}
\begin{document}
\begin{center}
	{\Large EGR 121: Engineering Innovation}\\
	{\Huge #(.Name)#}\\
	{\large \today}\\
\end{center}

\section*{Ratings}
\noindent
#(range $index, $plot := .Ratings)#\begin{minipage}[t]{0.49\linewidth}
	\centering
	\pgfplotstableread{#($plot.Data)#}{\data}
	\begin{tikzpicture}[scale=0.75,transform shape]
	\begin{axis}[ybar,
				 compat=newest,
				 bar width=32pt,
				 axis lines*=left,
				 ymin=0,ymax=#($plot.YMax)#,
				 y axis line style={opacity=0},
				 ytick={0,5,10,15,20,25,30,35,40,45,50},
				 yticklabels={\empty},
				 ytick style={draw=none},
				 xtick=data,
				 symbolic x coords={#($plot.RatingLabels)#},
				 x tick label style={rotate=90,anchor=east},
				 enlarge x limits=0.225,
				 nodes near coords,
				 title=#($plot.Title)#,
				 title style={yshift=-1.5em},
				 axis on top,
				 major grid style=white,
				 ymajorgrids,]
		\addplot[fill=black!60, draw=black!60] table [x={Rating}, y={Count}] {\data};
	\end{axis}
	\end{tikzpicture}
\end{minipage}#(if multipleOfTwo $index)#\\~\\#(else)#%#(end)#
#(end)#

\section*{Areas of Success}
#(if .PositiveFeedback)#\begin{itemize}
	#(range $index, $comment := .PositiveFeedback)#\item #($comment)#
	#(end)#
\end{itemize}
#(else)#No students gave feedback for this area.#(end)#

\section*{Areas of Improvement}
#(if .NegativeFeedback)#\begin{itemize}
	#(range $index, $comment := .NegativeFeedback)#\item #($comment)#
	#(end)#
\end{itemize}
#(else)#No students gave feedback for this area.#(end)#

\section*{Other Comments}
#(if .OtherFeedback)#\begin{itemize}
	#(range $index, $comment := .OtherFeedback)#\item #($comment)#
	#(end)#
\end{itemize}
#(else)#No students gave feedback for this area.#(end)#

\end{document}
`
)

// TAFeedbackData holds data for TeX template consumption
type TAFeedbackData struct {
	// Name is the name of the TA being rated
	Name string

	// Ratings holds the informaiton for plotting ratings
	Ratings []RatingPlot

	// PositiveFeedback is a list of feedback lines about what the TA did well
	PositiveFeedback []string

	// PositiveFeedback is a list of feedback lines about what the TA could improve on
	NegativeFeedback []string

	// PositiveFeedback is an optional list of feedback lines for other comments
	OtherFeedback []string
}

// RatingPlot holds nicely-formatted data for making a TeX plot using a template
type RatingPlot struct {
	// Title is the title of the plot
	Title string

	// YMax is the max Y value of the plot, this should be set to the total number of responses
	YMax int

	// Data is the formatted data table to be consumed by pgflots
	Data string

	// RatingLabels are the labels to be placed on the X axis of the plot
	RatingLabels string
}

// generateTeXFile generates a TeX file from a TAFeedback record and places it in the output directory
func generateTeXFile(record *TAFeedback, outputDirectory string) (string, error) {
	reportTemplate := template.New("report")
	reportTemplate = reportTemplate.Funcs(template.FuncMap{
		"multipleOfTwo": func(i int) bool {
			return i%2 == 1
		},
	})
	reportTemplate = reportTemplate.Delims("#(", ")#")
	reportTemplate, err := reportTemplate.Parse(reportOutline)
	if err != nil {
		return "", fmt.Errorf("error generating report template: %v", err)
	}

	outputFile := path.Join(outputDirectory, strings.Replace(record.Name, " ", "_", -1)+".tex")
	file, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("could not open destination file for report: %v", err)
	}
	defer file.Close()

	err = reportTemplate.Execute(file, convertToTeXData(record))
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return outputFile, nil
}

func convertToTeXData(input *TAFeedback) TAFeedbackData {
	var output TAFeedbackData

	output.Name = input.Name

	for criterion := range input.Ratings {
		title := strings.Title(criterion)

		ymax := 0

		labels := []string{"Very Poor", "Poor", "Adequate", "Good", "Very Good"}
		ratings := input.Ratings[criterion]
		var data bytes.Buffer

		// write header line to "csv"
		data.WriteString("Rating,Count\n")

		for i := 0; i < len(labels); i++ {
			// write data line to "csv"
			data.WriteString(fmt.Sprintf("%s,%d\n", labels[i], ratings[i]))
			ymax += ratings[i]
		}

		output.Ratings = append(output.Ratings, RatingPlot{
			Title:        title,
			YMax:         ymax,
			Data:         data.String(),
			RatingLabels: strings.Join(labels, ","),
		})
	}

	output.PositiveFeedback = input.PositiveFeedback
	output.NegativeFeedback = input.NegativeFeedback
	output.OtherFeedback = input.OtherFeedback

	return output
}
