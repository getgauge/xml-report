# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE in the project root for license information.
# ----------------------------------------------------------------

#!/bin/sh

#Using protoc version 3.0.0

cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc -I=. --go_out=. --go-grpc_out=. spec.proto messages.proto services.proto
mv github.com/getgauge/gauge-proto/go/gauge_messages/* ../gauge_messages
rm -rf github.com/