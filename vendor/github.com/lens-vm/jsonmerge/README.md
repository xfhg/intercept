# JSON Merge Patch
A `JSON Merge Patch` utility forked from [json-patch](https://github.com/evanphx/json-patch) and based on [fastjson](github.com/valyala/fastjson), with WASM/WASI support for [TinyGo](https://tinygo.org/).

## Wasm Support
The TinyGo compiler doesn't support `encoding/json` or some of the `reflect` package that most JSON serializers are based on.

Luckily `valyala/fastjson` is a native go implementation that doesn't rely on `encoding/json` and only uses the supported `reflect` functions.

Therefore this package is a TinyGo compliant library for `JSON Merge Patch`.

## Usage
Public Functions:
- `MergePatch(doc, merge) -> doc` : Applies a MergePatch to a JSON document and returns the new doc.
- `MergeMergePatch(merge, merge) -> merge` : Merges two Merge Patch documents into a single.

```go
package main

import (
	"fmt"

	"github.com/lens-vm/jsonmerge"
)

func main() {
	// Let's create a merge patch and a original document...
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	merge := []byte(`{"name": "Jane", "age": 21}`)

	new, err := jsonpatch.MergePatch(original, merge)
	if err != nil {
		panic(err)
	}

	fmt.Printf("new merged document:   %s\n", new)
	// outputs {{"name": "Jane", "age": 21, "height": 3.21}}
}
```

## Roadmap
Currently this project succesfully passes much of the [JSON Merge RFC Tests](https://tools.ietf.org/html/rfc7386) for applying merges to 1. Documents 2. Other Merges.

Next steps is to support creating a Merge from the diff of two existing JSON Docs. This is suppported in the orignal `evanphx/json-patch` library, but the changes haven't made its way here.

## IMPORTANT
This library is an adaptation of / based on [evanphx/json-patch](https://github.com/evanphx/json-patch). The goal of this fork is to replace the original libraries dependancy on `encoding/json` with `valyala/fastjson`. The original license (BSD 3-Clause) is maintained, and the original copyright is maintained in the unchanged files.

## Contributors
- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))
- Evan Pheonix ([@evanphx](https://github.com/evanphx)) (Original Fork)