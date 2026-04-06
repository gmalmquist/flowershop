package jsv

import (
	"fmt"
)

func ExampleJsonValue_Resolve() {
	blob := JsonValue(`{
    "phrases": {
      "greetings": {
        "salutations": "hello"
      }
    }
  }`)
	v, err := blob.Resolve("phrases.greetings.salutations")
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Print(v.String())
	// Output: hello
}

func ExampleJsonValue_ResolveOr() {
	blob := JsonValue(`{
    "phrases": {
      "greetings": {
        "salutations": "hello"
      }
    }
  }`)
	v := blob.ResolveOr("phrases.farewells.goodbye", JsonValue("bye :("))
	fmt.Print(v.String())
	// Output: bye :(
}
