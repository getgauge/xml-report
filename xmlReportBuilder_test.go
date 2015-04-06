// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/html-report.

// getgauge/html-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/html-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/html-report.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/xml"
	"github.com/getgauge/xml-report/gauge_messages"
	"github.com/golang/protobuf/proto"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestToVerifyXmlContent(c *C) {
	value := gauge_messages.ProtoItem_Scenario
	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{Failed: proto.Bool(false), ScenarioHeading: proto.String("Scenario1")}, ItemType: &value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: proto.String("HEADING"), FileName: proto.String("FILENAME"), Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: proto.Int(1), Failed: proto.Bool(false)}
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
	item := &gauge_messages.ProtoItem{Scenario: &gauge_messages.ProtoScenario{Failed: proto.Bool(true), ScenarioHeading: proto.String("Scenario1")}, ItemType: &value}
	spec := &gauge_messages.ProtoSpec{SpecHeading: proto.String("HEADING"), FileName: proto.String("FILENAME"), Items: []*gauge_messages.ProtoItem{item}}
	specResult := &gauge_messages.ProtoSpecResult{ProtoSpec: spec, ScenarioCount: proto.Int(1), Failed: proto.Bool(true), ScenarioFailedCount: proto.Int(1)}
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
	c.Assert(suites.Suites[0].TestCases[0].Failure.Message, Equals, "Failed")
	c.Assert(suites.Suites[0].TestCases[0].Failure.Contents, Equals, "")
}
