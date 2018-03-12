# godat - Simple data serializer for Go

[![Build Status](https://travis-ci.org/lokhman/godat.svg?branch=master)](https://travis-ci.org/lokhman/godat)
[![codecov](https://codecov.io/gh/lokhman/godat/branch/master/graph/badge.svg)](https://codecov.io/gh/lokhman/godat)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

	go get github.com/lokhman/godat

## Usage

    package main
    
    import "github.com/lokhman/godat"
    
    func main() {
        serializedBytes, err := godat.Marshal(anyData)
        if err != nil {
            log.Fatal(err)
        }
        
        err = godat.Unmarshal(serializedBytes, &unserializedValue)
        if err != nil {
            log.Fatal(err)
        }
        
        // anyData == unserializedValue
	}
	
## Tests

Use `go test` for testing.

## License

Library is available under the MIT license. The included LICENSE file describes this in detail.
