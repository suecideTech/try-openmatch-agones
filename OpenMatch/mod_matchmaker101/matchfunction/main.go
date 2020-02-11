package main

import (
	"matchfunction/mmf"
)

// This tutorial implenents a basic Match Function that is hosted in the below
// configured port. You can also configure the Open Match QueryService endpoint
// with which the Match Function communicates to query the Tickets.

const (
	queryServiceAddress = "om-query.open-match.svc.cluster.local:50503" // Address of the QueryService endpoint.
	serverPort          = 50502                                         // The port for hosting the Match Function.
)

func main() {
	mmf.Start(queryServiceAddress, serverPort)
}
