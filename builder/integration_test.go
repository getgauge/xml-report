//go:build linux || darwin
// +build linux darwin

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package builder

import (
	"encoding/xml"
	"os"
	"path/filepath"

	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
	. "gopkg.in/check.v1"
)

var junitSchema *xsd.Schema

func init() {
	schema, err := os.ReadFile(filepath.Join("_testdata", "junit.xsd"))
	if err != nil {
		panic(err)
	}
	junitSchema, err = xsd.Parse(schema)
	if err != nil {
		panic(err)
	}
}

func (s *MySuite) TestToVerifyXmlContent(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario1"}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

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
	steps := []*gauge_messages.ProtoItem{{Step: step, ItemType: stepType}}

	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario1",
		ScenarioItems: steps, ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

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
	c.Assert(suites.Suites[0].TestCases[0].Failure.Message, Equals, "Step Execution Failure: 'something'")
	c.Assert(suites.Suites[0].TestCases[0].Failure.Contents, Equals, "nice little stacktrace")
}

func (s *MySuite) TestToVerifyXmlContentForMultipleFailuresInExecutionResult(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	stepType := gauge_messages.ProtoItem_Step
	result1 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "fail but don't stop", StackTrace: "nice little stacktrace", RecoverableError: true}
	result2 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "stop here", StackTrace: "very easy to trace"}
	step1 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result1}}
	step2 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result2}}
	steps := []*gauge_messages.ProtoItem{{Step: step1, ItemType: stepType}, {Step: step2, ItemType: stepType}}

	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario1",
		ScenarioItems: steps, ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}, ItemType: value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, _ := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
	c.Assert(suites.Suites[0].TestCases[0].Classname, Equals, "HEADING")
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario1")
	failure := `Step Execution Failure: 'fail but don't stop'
nice little stacktrace

Step Execution Failure: 'stop here'
very easy to trace`
	c.Assert(suites.Suites[0].TestCases[0].Failure.Message, Equals, "Multiple failures")
	c.Assert(suites.Suites[0].TestCases[0].Failure.Contents, Equals, failure)
}

func (s *MySuite) TestToVerifyXmlContentForMultipleFailuresWithNestedConcept(c *C) {
	scenType := gauge_messages.ProtoItem_Scenario
	stepType := gauge_messages.ProtoItem_Step
	cptType := gauge_messages.ProtoItem_Concept

	result1 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "fail but don't stop", StackTrace: "nice little stacktrace"}
	result2 := &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "continue on failure", StackTrace: "very easy to trace"}

	step1 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result1}}
	step2 := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: result2}}
	steps := []*gauge_messages.ProtoItem{{Step: step1, ItemType: stepType}, {Step: step2, ItemType: stepType}}

	cpt1 := &gauge_messages.ProtoItem{Concept: &gauge_messages.ProtoConcept{Steps: steps}, ItemType: cptType}
	cpt2 := &gauge_messages.ProtoItem{Concept: &gauge_messages.ProtoConcept{Steps: append(steps, cpt1)}, ItemType: cptType}

	scenario := &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario1", ScenarioItems: append(steps, cpt2), ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}
	scenItem := &gauge_messages.ProtoItem{Scenario: scenario, ItemType: scenType}

	spec := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{scenItem}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: 1, Failed: true, ScenarioFailedCount: 1}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, _ := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(len(suites.Suites[0].TestCases), Equals, 1)
	c.Assert(suites.Suites[0].TestCases[0].Classname, Equals, "HEADING")
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario1")
	failure := `Step Execution Failure: 'fail but don't stop'
nice little stacktrace

Step Execution Failure: 'continue on failure'
very easy to trace

Concept Execution Failure: 'fail but don't stop'
nice little stacktrace

Concept Execution Failure: 'continue on failure'
very easy to trace

Concept Execution Failure: 'fail but don't stop'
nice little stacktrace

Concept Execution Failure: 'continue on failure'
very easy to trace`
	c.Assert(suites.Suites[0].TestCases[0].Failure.Message, Equals, "Multiple failures")
	c.Assert(suites.Suites[0].TestCases[0].Failure.Contents, Equals, failure)
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
	tableItem := &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Table,
		Table: &gauge_messages.ProtoTable{
			Headers: &gauge_messages.ProtoTableRow{
				Cells: []string{"name", "age"},
			},
			Rows: []*gauge_messages.ProtoTableRow{
				{Cells: []string{"john", "20"}},
				{Cells: []string{"mike", "22"}},
			},
		},
	}

	value := gauge_messages.ProtoItem_TableDrivenScenario
	scenario1 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	scenario2 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	item1 := &gauge_messages.ProtoItem{TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenario: &scenario1, IsSpecTableDriven: true, TableRowIndex: 0}, ItemType: value}
	item2 := &gauge_messages.ProtoItem{TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenario: &scenario2, IsSpecTableDriven: true, TableRowIndex: 1}, ItemType: value}
	spec1 := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{tableItem, item1, item2}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec1, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

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
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario | SpecRow: 1: [name: john] [age: 20]")
	c.Assert(suites.Suites[0].TestCases[1].Name, Equals, "Scenario | SpecRow: 2: [name: mike] [age: 22]")
}

func (s *MySuite) TestToVerifyXmlContentForScenarioTableDrivenExecution(c *C) {
	scenarioTable := &gauge_messages.ProtoTable{
		Headers: &gauge_messages.ProtoTableRow{
			Cells: []string{"city", "country"},
		},
		Rows: []*gauge_messages.ProtoTableRow{
			{Cells: []string{"New York", "USA"}},
			{Cells: []string{"London", "UK"}},
		},
	}

	value := gauge_messages.ProtoItem_TableDrivenScenario
	scenario1 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	scenario2 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	item1 := &gauge_messages.ProtoItem{
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{
			Scenario:              &scenario1,
			IsScenarioTableDriven: true,
			ScenarioTableRowIndex: 0,
			ScenarioDataTable:     scenarioTable,
		},
		ItemType: value,
	}
	item2 := &gauge_messages.ProtoItem{
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{
			Scenario:              &scenario2,
			IsScenarioTableDriven: true,
			ScenarioTableRowIndex: 1,
			ScenarioDataTable:     scenarioTable,
		},
		ItemType: value,
	}
	spec1 := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item1, item2}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec1, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

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
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario | ScnRow: 1: [city: New York] [country: USA]")
	c.Assert(suites.Suites[0].TestCases[1].Name, Equals, "Scenario | ScnRow: 2: [city: London] [country: UK]")
}

func (s *MySuite) TestToVerifyXmlContentForBothSpecAndScenarioTableDrivenExecution(c *C) {
	specTableItem := &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Table,
		Table: &gauge_messages.ProtoTable{
			Headers: &gauge_messages.ProtoTableRow{
				Cells: []string{"name", "age"},
			},
			Rows: []*gauge_messages.ProtoTableRow{
				{Cells: []string{"john", "20"}},
				{Cells: []string{"mike", "22"}},
			},
		},
	}

	scenarioTable := &gauge_messages.ProtoTable{
		Headers: &gauge_messages.ProtoTableRow{
			Cells: []string{"city", "country"},
		},
		Rows: []*gauge_messages.ProtoTableRow{
			{Cells: []string{"New York", "USA"}},
			{Cells: []string{"London", "UK"}},
		},
	}

	value := gauge_messages.ProtoItem_TableDrivenScenario
	scenario1 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	scenario2 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	item1 := &gauge_messages.ProtoItem{
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{
			Scenario:              &scenario1,
			IsSpecTableDriven:     true,
			TableRowIndex:         0,
			IsScenarioTableDriven: true,
			ScenarioTableRowIndex: 0,
			ScenarioDataTable:     scenarioTable,
		},
		ItemType: value,
	}
	item2 := &gauge_messages.ProtoItem{
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{
			Scenario:              &scenario2,
			IsSpecTableDriven:     true,
			TableRowIndex:         1,
			IsScenarioTableDriven: true,
			ScenarioTableRowIndex: 1,
			ScenarioDataTable:     scenarioTable,
		},
		ItemType: value,
	}
	spec1 := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{specTableItem, item1, item2}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec1, ScenarioCount: 1, Failed: false}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

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
	c.Assert(suites.Suites[0].TestCases[0].Name, Equals, "Scenario | SpecRow: 1: [name: john] [age: 20] ScnRow: 1: [city: New York] [country: USA]")
	c.Assert(suites.Suites[0].TestCases[1].Name, Equals, "Scenario | SpecRow: 2: [name: mike] [age: 22] ScnRow: 2: [city: London] [country: UK]")
}

func (s *MySuite) TestToVerifyXmlContentForErroredSpec(c *C) {
	value := gauge_messages.ProtoItem_TableDrivenScenario
	scenario1 := gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	item1 := &gauge_messages.ProtoItem{TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenario: &scenario1, TableRowIndex: 1}, ItemType: value}
	spec1 := &gauge_messages.ProtoSpec{SpecHeading: "HEADING", FileName: "FILENAME", Items: []*gauge_messages.ProtoItem{item1}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec1, ScenarioCount: 1, Failed: true, Errors: []*gauge_messages.Error{{Type: gauge_messages.Error_PARSE_ERROR, Message: "message"}}}
	suiteResult := &gauge_messages.ProtoSuiteResult{SpecResults: []*gauge_messages.ProtoSpecResult{specResult}}
	message := &gauge_messages.SuiteExecutionResult{SuiteResult: suiteResult}

	builder := &XmlBuilder{currentId: 0}
	bytes, err := builder.GetXmlContent(message)

	assertXmlValidation(bytes, c)

	var suites JUnitTestSuites
	xml.Unmarshal(bytes, &suites)

	c.Assert(err, Equals, nil)
	c.Assert(*suites.Suites[0].TestCases[0].Failure, Equals, JUnitFailure{
		Message:  "Parse/Validation Errors",
		Type:     "Parse/Validation Errors",
		Contents: "[Parse Error] message",
	})
}

func assertXmlValidation(xml []byte, c *C) {
	doc, err := libxml2.Parse(xml)
	c.Assert(err, Equals, nil)
	err = junitSchema.Validate(doc)
	if err != nil {
		var errors []string
		for _, e := range err.(xsd.SchemaValidationError).Errors() {
			errors = append(errors, e.Error())
		}
		c.Assert(err, Equals, nil, Commentf(strings.Join(errors, "\n")))
	}
}
