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
	"testing"

	"path/filepath"

	"github.com/getgauge/xml-report/gauge_messages"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestGetSpecNameWhenHeadingIsPresent(c *C) {
	want := "heading"

	got := getSpecName(&gauge_messages.ProtoSpec{SpecHeading: "heading"})

	c.Assert(want, Equals, got)
}

func (s *MySuite) TestGetSpecNameWhenHeadingIsNotPresent(c *C) {
	want := "example.spec"

	got := getSpecName(&gauge_messages.ProtoSpec{FileName: filepath.Join("specs", "specs1", "example.spec")})

	c.Assert(want, Equals, got)
}

func (s *MySuite) TestHasParseErrors(c *C) {
	errors := []*gauge_messages.Error{
		{Type: gauge_messages.Error_PARSE_ERROR},
		{Type: gauge_messages.Error_VALIDATION_ERROR},
	}

	got := hasParseErrors(errors)

	c.Assert(true, Equals, got)
}

func (s *MySuite) TestHasParseErrorsWithNoErrors(c *C) {
	errors := []*gauge_messages.Error{}

	got := hasParseErrors(errors)

	c.Assert(false, Equals, got)
}

func (s *MySuite) TestHasParseErrorsWithOnyValidationErrors(c *C) {
	errors := []*gauge_messages.Error{
		{Type: gauge_messages.Error_VALIDATION_ERROR},
		{Type: gauge_messages.Error_VALIDATION_ERROR},
	}

	got := hasParseErrors(errors)

	c.Assert(false, Equals, got)
}

func (s *MySuite) TestGetErrorTestCase(c *C) {
	res := &gauge_messages.ProtoSpecResult{
		ProtoSpec: &gauge_messages.ProtoSpec{
			SpecHeading: "heading",
		},
		Errors: []*gauge_messages.Error{
			{
				Type:    gauge_messages.Error_PARSE_ERROR,
				Message: "parse error",
			},
			{
				Type:    gauge_messages.Error_VALIDATION_ERROR,
				Message: "validation error",
			},
		},
		ExecutionTime: int64(1),
	}

	want := JUnitTestCase{
		Classname: "heading",
		Name:      "heading",
		Time:      formatTime(int(1)),
		Failure: &JUnitFailure{
			Message:  "Parse/Validation Errors",
			Type:     "Parse/Validation Errors",
			Contents: "[Parse Error] parse error\n[Validation Error] validation error",
		},
	}

	got := getErrorTestCase(res)

	c.Assert(want, DeepEquals, got)
}
