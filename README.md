# GopherNet
[![Build Status](https://github.com/henvic/gophernet/workflows/Tests/badge.svg)](https://github.com/henvic/gophernet/actions?query=workflow%3ATests)

Use `go install github.com/henvic/gophernet` to install application.

`gophernet -help` can be used to list available flags.

Use the `-verbose` flag to turn on debug log-level.

Setting the `-update-status-ticker` flag value to a number greater than one minute will speed up the simulation.

## Example of usage
Run the program with the following command:

`go run -race . -verbose -update-status-ticker 3s -burrow-expiration 100m -report-status-ticker 1s`

Read the output file:

`tail -f gophernet-report.txt`

List burrows:

`curl -XGET "http://localhost:8080/burrows"`

Rent a burrow:

`curl -XPOST "http://localhost:8080/burrows/The%20Deep%20Den/rent"`
