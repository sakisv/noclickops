package main

import (
	_ "embed"
	"fmt"
	"log/slog"
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

	slog.Info(fmt.Sprintf("Saving html report to %v", filename))
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(html)
}
