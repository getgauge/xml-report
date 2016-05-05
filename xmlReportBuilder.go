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

package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

func (self *XmlBuilder) getXmlContent(executionSuiteResult *gauge_messages.SuiteExecutionResult) ([]byte, error) {
	suiteResult := executionSuiteResult.GetSuiteResult()
	self.suites = JUnitTestSuites{}
	for _, result := range suiteResult.GetSpecResults() {
		self.getSpecContent(result)
	}
	bytes, err := xml.MarshalIndent(self.suites, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (self *XmlBuilder) getSpecContent(result *gauge_messages.ProtoSpecResult) {
	self.currentId += 1
	hostName, err := os.Hostname()
	if err != nil {
		hostName = hostname
	}
	ts := self.getTestSuite(result, hostName)
	if result.GetProtoSpec().GetPreHookFailure() != nil || result.GetProtoSpec().GetPostHookFailure() != nil {
		ts.Failures += 1
	}
	for _, test := range result.GetProtoSpec().GetItems() {
		if test.GetItemType() == gauge_messages.ProtoItem_Scenario {
			self.getScenarioContent(result, test.GetScenario(), &ts)
		} else if test.GetItemType() == gauge_messages.ProtoItem_TableDrivenScenario {
			self.getTableDrivenScenarioContent(result, test.GetTableDrivenScenario(), &ts)
		}
	}
	self.suites.Suites = append(self.suites.Suites, ts)
}

func (self *XmlBuilder) getScenarioContent(result *gauge_messages.ProtoSpecResult, scenario *gauge_messages.ProtoScenario, ts *JUnitTestSuite) {
	testCase := JUnitTestCase{
		Classname: result.GetProtoSpec().GetSpecHeading(),
		Name:      scenario.GetScenarioHeading(),
		Time:      formatTime(int(scenario.GetExecutionTime())),
		Failure:   nil,
	}
	if scenario.GetFailed() {
		message, content := self.getFailure(scenario)
		testCase.Failure = &JUnitFailure{
			Message:  message,
			Type:     message,
			Contents: content,
		}
	} else if scenario.GetSkipped() {
		testCase.SkipMessage = &JUnitSkipMessage{
			Message: strings.Join(scenario.SkipErrors, "\n"),
		}
	}
	ts.TestCases = append(ts.TestCases, testCase)
}

func (self *XmlBuilder) getTableDrivenScenarioContent(result *gauge_messages.ProtoSpecResult, tableDriven *gauge_messages.ProtoTableDrivenScenario, ts *JUnitTestSuite) {
	for i, scenario := range tableDriven.GetScenarios() {
		ts.Tests += 1
		*scenario.ScenarioHeading += " " + strconv.Itoa(i)
		self.getScenarioContent(result, scenario, ts)
	}
	ts.Tests -= 1
}

func (self *XmlBuilder) getTestSuite(result *gauge_messages.ProtoSpecResult, hostName string) JUnitTestSuite {
	now := time.Now()
	formattedNow := fmt.Sprintf(timeStampFormat, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	systemError := SystemErr{}
	if result.GetScenarioSkippedCount() > 0 {
		systemError.Contents = fmt.Sprintf("Validation failed, %d Scenarios were skipped.", result.GetScenarioSkippedCount())
	}
	return JUnitTestSuite{
		Id:               int(self.currentId),
		Tests:            int(result.GetScenarioCount()),
		Failures:         int(result.GetScenarioFailedCount()),
		Time:             formatTime(int(result.GetExecutionTime())),
		Timestamp:        formattedNow,
		Name:             result.GetProtoSpec().GetSpecHeading(),
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

func (self *XmlBuilder) getFailure(test *gauge_messages.ProtoScenario) (string, string) {
	msg, content := self.getFailureFromExecutionResult(test.GetScenarioHeading(), test.GetPreHookFailure(), test.GetPostHookFailure(), nil, "Scenario ")
	return self.perform(msg, content, func(test *gauge_messages.ProtoScenario) (string, string) {
		msg, content = self.getFailureFromSteps(test.GetContexts())
		return self.perform(msg, content, func(test *gauge_messages.ProtoScenario) (string, string) {
			return self.getFailureFromSteps(test.GetScenarioItems())
		}, test)
	}, test)
}

func (self *XmlBuilder) perform(msg string, content string, predicate func(test *gauge_messages.ProtoScenario) (string, string), test *gauge_messages.ProtoScenario) (string, string) {
	if msg != "" {
		return msg, content
	}
	return predicate(test)
}

func (self *XmlBuilder) getFailureFromSteps(items []*gauge_messages.ProtoItem) (string, string) {
	for _, item := range items {
		msg, err := "", ""
		if item.GetItemType() == gauge_messages.ProtoItem_Step {
			msg, err = self.getFailureFromExecutionResult(item.GetStep().GetActualText(), item.GetStep().GetStepExecutionResult().GetPreHookFailure(),
				item.GetStep().GetStepExecutionResult().GetPostHookFailure(),
				item.GetStep().GetStepExecutionResult().GetExecutionResult(), "Step ")
		} else if item.GetItemType() == gauge_messages.ProtoItem_Concept {
			msg, err = self.getFailureFromExecutionResult("", nil, nil, item.GetConcept().GetConceptExecutionResult().GetExecutionResult(), "Concept ")
		}
		if msg != "" {
			return msg, err
		}
	}
	return "", ""
}

func (self *XmlBuilder) getFailureFromExecutionResult(name string, preHookFailure *gauge_messages.ProtoHookFailure, postHookFailure *gauge_messages.ProtoHookFailure, stepExecutionResult *gauge_messages.ProtoExecutionResult, prefix string) (string, string) {
	if len(name) > 0 {
		name = fmt.Sprintf("%s\n", name)
	}
	if preHookFailure != nil {
		return fmt.Sprintf("%s%s%s: '%s'", name, prefix, preHookFailureMsg, preHookFailure.GetErrorMessage()), preHookFailure.GetStackTrace()
	} else if postHookFailure != nil {
		return fmt.Sprintf("%s%s%s: '%s'", name, prefix, postHookFailureMsg, postHookFailure.GetErrorMessage()), postHookFailure.GetStackTrace()
	} else if stepExecutionResult != nil && stepExecutionResult.GetFailed() {
		return fmt.Sprintf("%s%s%s: '%s'", name, prefix, executionFailureMsg, stepExecutionResult.GetErrorMessage()), stepExecutionResult.GetStackTrace()
	}
	return "", ""
}

func formatTime(time int) string {
	return fmt.Sprintf("%.3f", float64(time)/1000.0)
}
