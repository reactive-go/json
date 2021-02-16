package json_test

import (
	"fmt"
	"github.com/reactive-go/json"
)

func ExampleScanner_Next() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	sc := json.NewScanner([]byte(input))
	for {
		tok := sc.Next()
		if len(tok) < 1 {
			break
		}
		fmt.Printf("%s\n", tok)
	}

	// Output:
	// {
	// "a"
	// :
	// 1
	// ,
	// "b"
	// :
	// 123.456
	// ,
	// "c"
	// :
	// [
	// null
	// ]
	// }
}
