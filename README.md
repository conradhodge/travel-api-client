# Travel API client

![build](https://github.com/conradhodge/travel-api-client/workflows/Build/badge.svg)

An API client to fetch travel times, written in [Go](https://golang.org/).

This uses the [Traveline NextBuses API](https://www.travelinedata.org.uk/traveline-open-data/nextbuses-api/).

## Install

```shell
go get -u github.com/conradhodge/travel-api-client
```

## Development

For development, please ensure the following dependencies are installed.

- [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)
- [Go](https://golang.org/) 1.16+

Then run the following command to install the local dependencies:

```shell
make run setup
```

Run the following to see all available commands.

```shell
make
```
