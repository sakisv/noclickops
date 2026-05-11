package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"
)

//go:embed assets/report_template.html
var reportTemplate string

const reportDataPlaceholder string = "{{ REPORT_DATA }}"
const reportFilenamePrefix string = "report"
const reportFilenameExtension string = "html"

func generateReport(jsonData string) {
	html := strings.ReplaceAll(reportTemplate, reportDataPlaceholder, jsonData)
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("%v_%v.%v", reportFilenamePrefix, timestamp, reportFilenameExtension)

	println("Saving html report to ", filename)
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(html)
}
