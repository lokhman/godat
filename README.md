# godat - Simple data serializer for Go

## Install

	go get github.com/lokhman/godat

## Usage

    package main
    
    import "github.com/lokhman/godat"
    
    func main() {
        serializedBytes, err := Marshal(anyData)
        if err != nil {
            log.Fatal(err)
        }
        
        err = Unmarshal(serializedBytes, &unserializedValue)
        if err != nil {
            log.Fatal(err)
        }
        
        // anyData == unserializedValue
	}
	
## Tests

`go test` is used for testing.

## License

Library is available under the MIT license. The included LICENSE file describes this in detail.
