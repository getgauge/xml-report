// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/xml-report.

// getgauge/xml-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/xml-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/xml-report.  If not, see <http://www.gnu.org/licenses/>.

package builder

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"strconv"

	"path/filepath"

	"github.com/getgauge/xml-report/gauge_messages"
)

const (
	hostname            = "HOSTNAME"
	timeStampFormat     = "%d-%02d-%02dT%02d:%02d:%02d"
	preHookFailureMsg   = "Pre Hook Failure"
	postHookFailureMsg  = "Post Hook Failure"
	executionFailureMsg = "Execution Failure"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName          xml.Name        `xml:"testsuite"`
	Id               int             `xml:"id,attr"`
	Tests            int             `xml:"tests,attr"`
	Failures         int             `xml:"failures,attr"`
	Package          string          `xml:"package,attr"`
	Time             string          `xml:"time,attr"`
	Timestamp        string          `xml:"timestamp,attr"`
	Name             string          `xml:"name,attr"`
	Errors           int             `xml:"errors,attr"`
	SkippedTestCount int             `xml:"skipped,attr,omitempty"`
	Hostname         string          `xml:"hostname,attr"`
	Properties       []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases        []JUnitTestCase `xml:"testcase"`
	SystemOutput     SystemOut
	SystemError      SystemErr
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

type SystemOut struct {
	XMLName  xml.Name `xml:"system-out"`
	Contents string   `xml:",chardata"`
}

type SystemErr struct {
	XMLName  xml.Name `xml:"system-err"`
	Contents string   `xml:",chardata"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type XmlBuilder struct {
	currentId int
	suites    JUnitTestSuites
}

func NewXmlBuilder(id int) *XmlBuilder {
	return &XmlBuilder{currentId: id}
}

type StepFailure struct {
	Message string
	Err     string
}

func (x *XmlBuilder) GetXmlContent(executionSuiteResult *gauge_messages.SuiteExecutionResult) ([]byte, error) {
	suiteResult := executionSuiteResult.GetSuiteResult()
	x.suites = JUnitTestSuites{}
	for _, result := range suiteResult.GetSpecResults() {
		x.getSpecContent(result)
	}
	bytes, err := xml.MarshalIndent(x.suites, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (x *XmlBuilder) getSpecContent(result *gauge_messages.ProtoSpecResult) {
	x.currentId += 1
	hostName, err := os.Hostname()
	if err != nil {
		hostName = hostname
	}
	ts := x.getTestSuite(result, hostName)
	if hasParseErrors(result.Errors) {
		ts.Failures++
		ts.TestCases = append(ts.TestCases, getErrorTestCase(result))
	} else {
		s := result.GetProtoSpec()
		ts.Failures += len(s.GetPreHookFailures()) + len(s.GetPostHookFailures())
		for _, test := range result.GetProtoSpec().GetItems() {
			if test.GetItemType() == gauge_messages.ProtoItem_Scenario {
				x.getScenarioContent(result, test.GetScenario(), &ts)
			} else if test.GetItemType() == gauge_messages.ProtoItem_TableDrivenScenario {
				x.getTableDrivenScenarioContent(result, test.GetTableDrivenScenario(), &ts)
			}
		}
	}
	x.suites.Suites = append(x.suites.Suites, ts)
}
func getErrorTestCase(result *gauge_messages.ProtoSpecResult) JUnitTestCase {
	var failures []string
	for _, e := range result.Errors {
		t := "Parse"
		if e.Type == gauge_messages.Error_VALIDATION_ERROR {
			t = "Validation"
		}
		failures = append(failures, fmt.Sprintf("[%s Error] %s", t, e.Message))
	}
	return JUnitTestCase{
		Classname: getSpecName(result.GetProtoSpec()),
		Name:      getSpecName(result.GetProtoSpec()),
		Time:      formatTime(int(result.GetExecutionTime())),
		Failure: &JUnitFailure{
			Message:  "Parse/Validation Errors",
			Type:     "Parse/Validation Errors",
			Contents: strings.Join(failures, "\n"),
		},
	}
}

func (x *XmlBuilder) getScenarioContent(result *gauge_messages.ProtoSpecResult, scenario *gauge_messages.ProtoScenario, ts *JUnitTestSuite) {
	testCase := JUnitTestCase{
		Classname: getSpecName(result.GetProtoSpec()),
		Name:      scenario.GetScenarioHeading(),
		Time:      formatTime(int(scenario.GetExecutionTime())),
		Failure:   nil,
	}
	if scenario.GetFailed() {
		var errors []string
		failures := x.getFailure(scenario)
		message := "Multiple failures"
		for _, step := range failures {
			errors = append(errors, fmt.Sprintf("%s\n%s", step.Message, step.Err))
		}
		if len(failures) == 1 {
			message = failures[0].Message
			errors = []string{failures[0].Err}
		}
		testCase.Failure = &JUnitFailure{
			Message:  message,
			Type:     message,
			Contents: strings.Join(errors, "\n\n"),
		}
	} else if scenario.GetSkipped() {
		testCase.SkipMessage = &JUnitSkipMessage{
			Message: strings.Join(scenario.SkipErrors, "\n"),
		}
	}
	ts.TestCases = append(ts.TestCases, testCase)
}

func (x *XmlBuilder) getTableDrivenScenarioContent(result *gauge_messages.ProtoSpecResult, tableDriven *gauge_messages.ProtoTableDrivenScenario, ts *JUnitTestSuite) {
	if tableDriven.GetScenario() != nil {
		scenario := tableDriven.GetScenario()
		scenario.ScenarioHeading += " " + strconv.Itoa(int(tableDriven.GetTableRowIndex())+1)
		x.getScenarioContent(result, scenario, ts)
	}
}

func (x *XmlBuilder) getTestSuite(result *gauge_messages.ProtoSpecResult, hostName string) JUnitTestSuite {
	now := time.Now()
	formattedNow := fmt.Sprintf(timeStampFormat, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	systemError := SystemErr{}
	if result.GetScenarioSkippedCount() > 0 {
		systemError.Contents = fmt.Sprintf("Validation failed, %d Scenarios were skipped.", result.GetScenarioSkippedCount())
	}
	return JUnitTestSuite{
		Id:               int(x.currentId),
		Tests:            int(result.GetScenarioCount()),
		Failures:         int(result.GetScenarioFailedCount()),
		Time:             formatTime(int(result.GetExecutionTime())),
		Timestamp:        formattedNow,
		Name:             getSpecName(result.GetProtoSpec()),
		Errors:           0,
		Hostname:         hostName,
		Package:          result.GetProtoSpec().GetFileName(),
		Properties:       []JUnitProperty{},
		TestCases:        []JUnitTestCase{},
		SkippedTestCount: int(result.GetScenarioSkippedCount()),
		SystemOutput:     SystemOut{},
		SystemError:      systemError,
	}
}

func (x *XmlBuilder) getFailure(test *gauge_messages.ProtoScenario) []StepFailure {
	errInfo := []StepFailure{}
	hookInfo := x.getFailureFromExecutionResult(test.GetScenarioHeading(), test.GetPreHookFailure(), test.GetPostHookFailure(), nil, "Scenario ")
	if hookInfo.Message != "" {
		return append(errInfo, hookInfo)
	}
	contextsInfo := x.getFailureFromSteps(test.GetContexts(), "Step ")
	if len(contextsInfo) > 0 {
		errInfo = append(errInfo, contextsInfo...)
	}
	stepsInfo := x.getFailureFromSteps(test.GetScenarioItems(), "Step ")
	if len(stepsInfo) > 0 {
		errInfo = append(errInfo, stepsInfo...)
	}
	return errInfo
}

func (x *XmlBuilder) getFailureFromSteps(items []*gauge_messages.ProtoItem, prefix string) []StepFailure {
	errInfo := []StepFailure{}
	for _, item := range items {
		stepInfo := StepFailure{Message: "", Err: ""}
		if item.GetItemType() == gauge_messages.ProtoItem_Step {
			preHookFailure := item.GetStep().GetStepExecutionResult().GetPreHookFailure()
			postHookFailure := item.GetStep().GetStepExecutionResult().GetPostHookFailure()
			result := item.GetStep().GetStepExecutionResult().GetExecutionResult()
			stepInfo = x.getFailureFromExecutionResult(item.GetStep().GetActualText(), preHookFailure, postHookFailure, result, prefix)
		} else if item.GetItemType() == gauge_messages.ProtoItem_Concept {
			errInfo = append(errInfo, x.getFailureFromSteps(item.GetConcept().GetSteps(), "Concept ")...)
		}
		if stepInfo.Message != "" {
			errInfo = append(errInfo, stepInfo)
		}
	}
	return errInfo
}

func (x *XmlBuilder) getFailureFromExecutionResult(name string, preHookFailure *gauge_messages.ProtoHookFailure,
	postHookFailure *gauge_messages.ProtoHookFailure, stepExecutionResult *gauge_messages.ProtoExecutionResult, prefix string) StepFailure {
	if len(name) > 0 {
		name = fmt.Sprintf("%s\n", name)
	}
	if preHookFailure != nil {
		return StepFailure{Message: fmt.Sprintf("%s%s%s: '%s'", name, prefix, preHookFailureMsg, preHookFailure.GetErrorMessage()), Err: preHookFailure.GetStackTrace()}
	} else if postHookFailure != nil {
		return StepFailure{Message: fmt.Sprintf("%s%s%s: '%s'", name, prefix, postHookFailureMsg, postHookFailure.GetErrorMessage()), Err: postHookFailure.GetStackTrace()}
	} else if stepExecutionResult != nil && stepExecutionResult.GetFailed() {
		return StepFailure{Message: fmt.Sprintf("%s%s%s: '%s'", name, prefix, executionFailureMsg, stepExecutionResult.GetErrorMessage()), Err: stepExecutionResult.GetStackTrace()}
	}
	return StepFailure{"", ""}
}

func getSpecName(spec *gauge_messages.ProtoSpec) string {
	if strings.TrimSpace(spec.SpecHeading) == "" {
		return filepath.Base(spec.GetFileName())
	}
	return spec.SpecHeading
}

func hasParseErrors(errors []*gauge_messages.Error) bool {
	for _, e := range errors {
		if e.Type == gauge_messages.Error_PARSE_ERROR {
			return true
		}
	}
	return false
}

func formatTime(time int) string {
	return fmt.Sprintf("%.3f", float64(time)/1000.0)
}
