# godat - Simple data serializer for Go 1.7+

[![Build Status](https://travis-ci.org/lokhman/godat.svg?branch=master)](https://travis-ci.org/lokhman/godat)
[![codecov](https://codecov.io/gh/lokhman/godat/branch/master/graph/badge.svg)](https://codecov.io/gh/lokhman/godat)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

	go get github.com/lokhman/godat

## Usage

    import "github.com/lokhman/godat"
    
    func main() {
        var anyData = ...
        var unserializedData ...
    
        serializedBytes, err := godat.Marshal(anyData)
        if err != nil {
            panic(err)
        }
        
        err = godat.Unmarshal(serializedBytes, &unserializedData)
        if err != nil {
            panic(err)
        }
        
        // anyData == unserializedData
	}
	
## Tests

Use `go test` for testing.

## License

Library is available under the MIT license. The included LICENSE file describes this in detail.
