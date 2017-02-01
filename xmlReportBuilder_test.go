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
	"testing"

	"github.com/getgauge/xml-report/gauge_messages"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestToVerifyXmlContent(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{Failed: false, ScenarioHeading: "Scenario1"}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.getXmlContent(message)
	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(err, Equals, nil)
	c.Assert(len(suites.Suites), Equals, 1)
	c.Assert(suites.Suites[0].Errors, Equals, 0)
	c.Assert(suites.Suites[0].Failures, Equals, 0)
	c.Assert(suites.Suites[0].Package, Equals, "FILENAME")
	c.Assert(suites.Suites[0].Name, Equals, "HEADING")
	c.Assert(suites.Suites[0].Tests, Equals, 1)
	c.Assert(suites.Suites[0].Timestamp, Equals, builder.suites.Suites[0].Timestamp)
	c.Assert(suites.Suites[0].SystemError.Contents, Equals, "")
	c.Assert(suites.Suites[0].SystemOutput.Contents, Equals, "")
	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
}

func (s *MySuite) TestToVerifyXmlContentForFailingExecutionResult(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	stepType := gauge_messages.ProtoItem_Step
	result := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "something", StackTrace: "nice little stacktrace"}
	step := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result}}
	steps := []*gauge_messages.ProtoItem{&gauge_messages.ProtoItem{Step: step, ItemType: stepType}}

	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{Failed: true,
		ScenarioHeading: "Scenario1", ScenarioItems: steps}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.getXmlContent(message)
	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(err, Equals, nil)
	c.Assert(len(suites.Suites), Equals, 1)
	// spec1 || testSuite
	c.Assert(suites.Suites[0].Errors, Equals, 0)
	c.Assert(suites.Suites[0].Failures, Equals, 1)
	c.Assert(suites.Suites[0].Package, Equals, "FILENAME")
	c.Assert(suites.Suites[0].Name, Equals, "HEADING")
	c.Assert(suites.Suites[0].Tests, Equals, 1)
	c.Assert(suites.Suites[0].Timestamp, Equals, builder.suites.Suites[0].Timestamp)
	c.Assert(suites.Suites[0].SystemError.Contents, Equals, "")
	c.Assert(suites.Suites[0].SystemOutput.Contents, Equals, "")
	// scenario1 of spec1 || testCase
	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
	c.Assert(suites.Suites[0].TestCases[0].Classname, Equals, "HEADING")
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario1")
	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Message, Equals, "Step Execution Failure: 'something'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Contents, Equals, "nice little stacktrace")
}

func (s *MySuite) TestToVerifyXmlContentForMultipleFailuresInExecutionResult(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	stepType := gauge_messages.ProtoItem_Step
	result1 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "fail but don't stop", StackTrace: "nice little stacktrace", RecoverableError: true}
	result2 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "stop here", StackTrace: "very easy to trace"}
	step1 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result1}}
	step2 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result2}}
	steps := []*gauge_messages.ProtoItem{&gauge_messages.ProtoItem{Step: step1, ItemType: stepType}, &gauge_messages.ProtoItem{Step: step2, ItemType: stepType}}

	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{Failed: true,
		ScenarioHeading: "Scenario1", ScenarioItems: steps}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, _ := builder.getXmlContent(message)
	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
	c.Assert(suites.Suites[0].TestCases[0].Classname, Equals, "HEADING")
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario1")
	c.Assert(len(suites.Suites[0].TestCases[0].Failures), Equals, 2)
	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Message, Equals, "Step Execution Failure: 'fail but don't stop'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Contents, Equals, "nice little stacktrace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[1].Message, Equals, "Step Execution Failure: 'stop here'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[1].Contents, Equals, "very easy to trace")
}

func (s *MySuite) TestToVerifyXmlContentForMultipleFailuresWithNestedConcept(c *C) {
	scenType := gauge_messages.ProtoItem_Scenario
	stepType := gauge_messages.ProtoItem_Step
	cptType := gauge_messages.ProtoItem_Concept

	result1 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "fail but don't stop", StackTrace: "nice little stacktrace"}
	result2 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "continue on failure", StackTrace: "very easy to trace"}

	step1 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result1}}
	step2 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result2}}
	steps := []*gauge_messages.ProtoItem{&gauge_messages.ProtoItem{Step: step1, ItemType: stepType}, &gauge_messages.ProtoItem{Step: step2, ItemType: stepType}}

	cpt1 := &gauge_messages.ProtoItem{Concept: &gauge_messages.ProtoConcept{Steps: steps}, ItemType: cptType}
	cpt2 := &gauge_messages.ProtoItem{Concept: &gauge_messages.ProtoConcept{Steps: append(steps, cpt1)}, ItemType: cptType}

	scenario := &gauge_messages.ProtoScenario{Failed: true, ScenarioHeading: "Scenario1", ScenarioItems: append(steps, cpt2)}
	scenItem := &gauge_messages.ProtoItem{Scenario: scenario, ItemType: scenType}

	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{scenItem}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, _ := builder.getXmlContent(message)
	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)
	for _, f := range suites.Suites[0].TestCases[0].Failures {
		println(f.Message)
	}
	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
	c.Assert(suites.Suites[0].TestCases[0].Classname, Equals, "HEADING")
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario1")
	c.Assert(len(suites.Suites[0].TestCases[0].Failures), Equals, 6)

	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Message, Equals, "Step Execution Failure: 'fail but don't stop'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[0].Contents, Equals, "nice little stacktrace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[1].Message, Equals, "Step Execution Failure: 'continue on failure'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[1].Contents, Equals, "very easy to trace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[2].Message, Equals, "Concept Execution Failure: 'fail but don't stop'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[2].Contents, Equals, "nice little stacktrace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[3].Message, Equals, "Concept Execution Failure: 'continue on failure'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[3].Contents, Equals, "very easy to trace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[4].Message, Equals, "Concept Execution Failure: 'fail but don't stop'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[4].Contents, Equals, "nice little stacktrace")
	c.Assert(suites.Suites[0].TestCases[0].Failures[5].Message, Equals, "Concept Execution Failure: 'continue on failure'")
	c.Assert(suites.Suites[0].TestCases[0].Failures[5].Contents, Equals, "very easy to trace")
}

func (s *MySuite) TestToVerifyXmlContentForFailingHookExecutionResult(c *C) {
	builder := &XmlBuilder{currentId: 0}
	info := builder.getFailureFromExecutionResult("", nil, nil, nil, "PREFIX ")

	c.Assert(info.Message, Equals, "")
	c.Assert(info.Err, Equals, "")

	failure := &gauge_messages.ProtoHookFailure{StackTrace: "StackTrace", ErrorMessage: "ErrorMessage"}
	hookInfo := builder.getFailureFromExecutionResult("", failure, nil, nil, "PREFIX ")

	c.Assert(hookInfo.Message, Equals, "PREFIX "+preHookFailureMsg+": 'ErrorMessage'")
	c.Assert(hookInfo.Err, Equals, "StackTrace")

	hookInfo = builder.getFailureFromExecutionResult("", nil, failure, nil, "PREFIX ")

	c.Assert(hookInfo.Message, Equals, "PREFIX "+postHookFailureMsg+": 'ErrorMessage'")
	c.Assert(hookInfo.Err, Equals, "StackTrace")

	hookInfo = builder.getFailureFromExecutionResult("Foo", nil, failure, nil, "PREFIX ")

	c.Assert(hookInfo.Message, Equals, "Foo\nPREFIX "+postHookFailureMsg+": 'ErrorMessage'")
	c.Assert(hookInfo.Err, Equals, "StackTrace")

	executionFailure := &gauge_messages.ProtoExecutionResult{StackTrace: "StackTrace", ErrorMessage: "ErrorMessage", Failed: true}
	execInfo := builder.getFailureFromExecutionResult("Foo", nil, nil, executionFailure, "PREFIX ")

	c.Assert(execInfo.Message, Equals, "Foo\nPREFIX "+executionFailureMsg+": 'ErrorMessage'")
	c.Assert(execInfo.Err, Equals, "StackTrace")
}

func (s *MySuite) TestToVerifyXmlContentForDataTableDrivenExecution(c *C) {
	value := gauge_messages.ProtoItem_TableDrivenScenario
	scenario1 := gauge_messages.ProtoScenario{Failed: false, ScenarioHeading: "Scenario"}
	scenario2 := gauge_messages.ProtoScenario{Failed: false, ScenarioHeading: "Scenario"}
	item1 := &gauge_messages.ProtoItem{TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenario: &scenario1, TableRowIndex: 1}, ItemType: value}
	item2 := &gauge_messages.ProtoItem{TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenario: &scenario2, TableRowIndex: 2}, ItemType: value}
	spec1 := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item1, item2}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec1, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.getXmlContent(message)
	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(err, Equals, nil)
	c.Assert(len(suites.Suites), Equals, 1)
	c.Assert(suites.Suites[0].Errors, Equals, 0)
	c.Assert(suites.Suites[0].Failures, Equals, 0)
	c.Assert(suites.Suites[0].Package, Equals, "FILENAME")
	c.Assert(suites.Suites[0].Name, Equals, "HEADING")
	c.Assert(suites.Suites[0].Tests, Equals, 1)
	c.Assert(suites.Suites[0].Timestamp, Equals, builder.suites.Suites[0].Timestamp)
	c.Assert(suites.Suites[0].SystemError.Contents, Equals, "")
	c.Assert(suites.Suites[0].SystemOutput.Contents, Equals, "")
	c.Assert(len(suites.Suites[0].TestCases), Equals, 2)
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario 2")
	c.Assert(suites.Suites[0].TestCases[1].Name, Equals, "Scenario 3")
}
