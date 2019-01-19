package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

//JUnitReport represents either a single test suite or a collection of test suites
type JUnitReport struct {
	XMLName   xml.Name
	XML       string       `xml:",innerxml"`
	Name      string       `xml:"name,attr"`
	Time      float64      `xml:"time,attr"`
	Tests     uint64       `xml:"tests,attr"`
	Failures  uint64       `xml:"failures,attr"`
	XMLBuffer bytes.Buffer `xml:"-"`
}

var usage = `Usage: junit-merger [options] [files]

Options:
  -o  Merged report filename`

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
	outputFileName := flag.String("o", "", "merged report filename")
	flag.Parse()
	files := flag.Args()
	printReport := *outputFileName == ""

	if len(files) == 0 {
		flag.Usage()
		return
	}

	var mergedReport JUnitReport
	fileCount := 0

	for _, fileName := range files {
		var report JUnitReport
		in, err := ioutil.ReadFile(fileName)
		if err != nil {
			panic(err)
		}

		err = xml.Unmarshal(in, &report)
		if err != nil {
			panic(err)
		}

		if report.XMLName.Local == "testsuites" {
			panic(errors.New("Reports with a root <testsuites> are not supported"))
		}

		fileCount++
		mergedReport.XMLName = xml.Name{Local: "testsuite"}
		mergedReport.Name = report.Name
		mergedReport.Time += report.Time
		mergedReport.Tests += report.Tests
		mergedReport.Failures += report.Failures
		mergedReport.XMLBuffer.WriteString(report.XML)
	}

	mergedReport.XML = mergedReport.XMLBuffer.String()
	mergedOutput, _ := xml.MarshalIndent(&mergedReport, "", "  ")

	if printReport {
		fmt.Println(string(mergedOutput))
	} else {
		os.MkdirAll(filepath.Dir(*outputFileName), os.ModePerm)

		err := ioutil.WriteFile(*outputFileName, mergedOutput, 0644)
		if err != nil {
			panic(err)
		}

		fmt.Println("Merged " + strconv.Itoa(fileCount) + " reports to " + *outputFileName)
	}
}
