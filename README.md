# A golang limit timed reader.

It generates output either buffer is full or time interval is reached.

## Installation

go get -d -v github.com/SeTriones/limitreader

## Example

```go
package main

import (
	"github.com/SeTriones/limitreader"
)

func main() {
	// the first paramter is max buf size, the second is time interval (in milliseocnds.)
	// reader will transfer its input items to output either on buff is full or time passes interval milliseocnds since last transfer.
	r := NewReader(200, 50)

	// output is a readonly channel for consumers to read
	output := r.GetOutput()	

	for i := 0; i < 5; i++ {
		go func(r *Reader) {
			for values := range output {
				// your code here
			}
		}(r)
	}		

	for i := 0; i < 100000; i++ {
		r.Put(i)
	}
}
```
