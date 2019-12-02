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
	"net"
	"os"

	"github.com/getgauge/xml-report/logger"

	gm "github.com/getgauge/xml-report/gauge_messages"
	"google.golang.org/grpc"
)

func main() {
	findPluginAndProjectRoot()
	if os.Getenv(pluginActionEnv) == executionAction {
		os.Chdir(projectRoot)
		address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			logger.Fatal("failed to start server.")
		}
		l, err := net.ListenTCP("tcp", address)
		if err != nil {
			logger.Fatal("failed to start server.")
		}
		server := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 1024 * 10))
		h := &handler{server: server}
		gm.RegisterReporterServer(server, h)
		logger.Info("Listening on port:%d", l.Addr().(*net.TCPAddr).Port)
		server.Serve(l)
	}
}
