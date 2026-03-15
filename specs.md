# Demo application for interacting with Tigerbeetle

I want to build an application that
* is written in Go
* interacts with Tigerbeetle
* simulates humans selling and buying stuff from each other, for random amounts, at random intervals
* and some sort of simulated central bank that provides an regular universal basic income to the agents.


## Application config
Every application generates N simulated humans (N being a command line parameter), starting at some ID offset (IDoffset, cli parameter). The application checks tigerbeetle for the existence of the accounts, and if not, creates them.

For now, the application has no metrics, just an occassional message on stdout with some basic information, like the credit on each simulated agent.
SIGINT is just to terminate the application.

Tigerbeetle is running on localhost:3000
