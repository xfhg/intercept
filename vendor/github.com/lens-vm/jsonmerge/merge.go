/*
 * Copyright (c) 2021, John-Alan Simmons
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. Neither the name of mosquitto nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package jsonmerge

// This file is copied from sections of
// https://github.com/evanphx/json-patch/blob/07737475986b62bc0ac8681f972a3c4ff771cef6/merge.go
// and modified to use github.com/valyala/fastjson instead of encoding/json

import (
	"fmt"

	"github.com/valyala/fastjson"
)

func merge(cur, patch *fastjson.Value, mergeMerge bool) *fastjson.Value {
	if _, err := cur.Object(); err != nil {
		pruneNulls(patch)
		return patch
	}

	if _, err := patch.Object(); err != nil {
		return patch
	}

	mergeDocs(cur, patch, mergeMerge)

	return cur
}

func mergeDocs(doc, patch *fastjson.Value, mergeMerge bool) error {
	// docObj, err := doc.Object()
	// if err != nil {
	// 	return err
	// }
	patchObj, err := patch.Object()
	if err != nil {
		return err
	}
	patchObj.Visit(func(key []byte, v *fastjson.Value) {
		k := string(key)
		if v.Type() == fastjson.TypeNull {
			if mergeMerge {
				doc.Set(k, v)
			} else {
				doc.Del(k)
			}
		} else {
			// cur, ok := (*doc)[k]
			cur := doc.Get(k)

			if cur == nil || cur.Type() == fastjson.TypeNull {
				pruneNulls(v)
				doc.Set(k, v)
			} else {
				// (*doc)[k] = merge(cur, v, mergeMerge)
				doc.Set(k, merge(cur, v, mergeMerge))
			}
		}
	})
	return nil
}

func pruneNulls(n *fastjson.Value) error {
	var err error
	if n.Type() == fastjson.TypeObject {
		_, err = pruneDocNulls(n)
	} else if n.Type() == fastjson.TypeArray {
		ary, err := n.Array()

		if err == nil {
			_, err = pruneAryNulls(&ary)
		}
	}

	return err
}

func pruneDocNulls(doc *fastjson.Value) (*fastjson.Value, error) {
	docObj, err := doc.Object()
	if err != nil {
		return nil, err
	}
	docObj.Visit(func(key []byte, v *fastjson.Value) {
		k := string(key)
		if v.Type() == fastjson.TypeNull {
			docObj.Del(k)
		} else {
			pruneNulls(v)
		}
	})

	return doc, nil
}

func pruneAryNulls(ary *[]*fastjson.Value) (*[]*fastjson.Value, error) {
	newAry := []*fastjson.Value{}

	for _, v := range *ary {
		if v != nil {
			err := pruneNulls(v)
			if err != nil {
				return nil, err
			}
			newAry = append(newAry, v)
		}
	}

	*ary = newAry

	return ary, nil
}

// var ErrBadJSONDoc = fmt.Errorf("Invalid JSON Document")
var ErrBadJSONPatch = fmt.Errorf("Invalid JSON Patch")

// var errBadMergeTypes = fmt.Errorf("Mismatched JSON Documents")

// MergeMergePatches merges two merge patches together, such that
// applying this resulting merged merge patch to a document yields the same
// as merging each merge patch to the document in succession.
func MergeMergePatches(patch1Data, patch2Data []byte) ([]byte, error) {
	return doMergePatch(patch1Data, patch2Data, true)
}

// MergePatch merges the patchData into the docData.
func MergePatch(docData, patchData []byte) ([]byte, error) {
	return doMergePatch(docData, patchData, false)
}

func doMergePatch(docData, patchData []byte, mergeMerge bool) ([]byte, error) {
	doc, err := fastjson.ParseBytes(docData)
	if err != nil {
		return nil, err
	}

	patch, err := fastjson.ParseBytes(patchData)
	if err != nil {
		return nil, err
	}

	_, docErr := doc.Object()
	_, patchErr := patch.Object()

	if docErr != nil || patchErr != nil {
		// Not an error, just not a doc, so we turn straight into the patch
		if patchErr == nil {
			if mergeMerge {
				doc = patch
			} else {
				doc, err = pruneDocNulls(patch)
			}
		} else {
			patchAry, patchErr := patch.Array()

			if patchErr != nil {
				return nil, ErrBadJSONPatch
			}

			pruneAryNulls(&patchAry)

			out := patch.MarshalTo(nil)

			// if patchErr != nil {
			// 	return nil, ErrBadJSONPatch
			// }

			return out, nil
		}
	} else {
		mergeDocs(doc, patch, mergeMerge)
	}

	return doc.MarshalTo(nil), nil
}

// // resemblesJSONArray indicates whether the byte-slice "appears" to be
// // a JSON array or not.
// // False-positives are possible, as this function does not check the internal
// // structure of the array. It only checks that the outer syntax is present and
// // correct.
// func resemblesJSONArray(input []byte) bool {
// 	input = bytes.TrimSpace(input)

// 	hasPrefix := bytes.HasPrefix(input, []byte("["))
// 	hasSuffix := bytes.HasSuffix(input, []byte("]"))

// 	return hasPrefix && hasSuffix
// }

// // CreateMergePatch will return a merge patch document capable of converting
// // the original document(s) to the modified document(s).
// // The parameters can be bytes of either two JSON Documents, or two arrays of
// // JSON documents.
// // The merge patch returned follows the specification defined at http://tools.ietf.org/html/draft-ietf-appsawg-json-merge-patch-07
// func CreateMergePatch(originalJSON, modifiedJSON []byte) ([]byte, error) {
// 	originalResemblesArray := resemblesJSONArray(originalJSON)
// 	modifiedResemblesArray := resemblesJSONArray(modifiedJSON)

// 	// Do both byte-slices seem like JSON arrays?
// 	if originalResemblesArray && modifiedResemblesArray {
// 		return createArrayMergePatch(originalJSON, modifiedJSON)
// 	}

// 	// Are both byte-slices are not arrays? Then they are likely JSON objects...
// 	if !originalResemblesArray && !modifiedResemblesArray {
// 		return createObjectMergePatch(originalJSON, modifiedJSON)
// 	}

// 	// None of the above? Then return an error because of mismatched types.
// 	return nil, errBadMergeTypes
// }

// // createObjectMergePatch will return a merge-patch document capable of
// // converting the original document to the modified document.
// func createObjectMergePatch(originalJSON, modifiedJSON []byte) ([]byte, error) {
// 	originalDoc := map[string]interface{}{}
// 	modifiedDoc := map[string]interface{}{}

// 	err := json.Unmarshal(originalJSON, &originalDoc)
// 	if err != nil {
// 		return nil, ErrBadJSONDoc
// 	}

// 	err = json.Unmarshal(modifiedJSON, &modifiedDoc)
// 	if err != nil {
// 		return nil, ErrBadJSONDoc
// 	}

// 	dest, err := getDiff(originalDoc, modifiedDoc)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return json.Marshal(dest)
// }

// // createArrayMergePatch will return an array of merge-patch documents capable
// // of converting the original document to the modified document for each
// // pair of JSON documents provided in the arrays.
// // Arrays of mismatched sizes will result in an error.
// func createArrayMergePatch(originalJSON, modifiedJSON []byte) ([]byte, error) {
// 	originalDocs := []json.RawMessage{}
// 	modifiedDocs := []json.RawMessage{}

// 	err := json.Unmarshal(originalJSON, &originalDocs)
// 	if err != nil {
// 		return nil, ErrBadJSONDoc
// 	}

// 	err = json.Unmarshal(modifiedJSON, &modifiedDocs)
// 	if err != nil {
// 		return nil, ErrBadJSONDoc
// 	}

// 	total := len(originalDocs)
// 	if len(modifiedDocs) != total {
// 		return nil, ErrBadJSONDoc
// 	}

// 	result := []json.RawMessage{}
// 	for i := 0; i < len(originalDocs); i++ {
// 		original := originalDocs[i]
// 		modified := modifiedDocs[i]

// 		patch, err := createObjectMergePatch(original, modified)
// 		if err != nil {
// 			return nil, err
// 		}

// 		result = append(result, json.RawMessage(patch))
// 	}

// 	return json.Marshal(result)
// }

// // Returns true if the array matches (must be json types).
// // As is idiomatic for go, an empty array is not the same as a nil array.
// func matchesArray(a, b []interface{}) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}
// 	if (a == nil && b != nil) || (a != nil && b == nil) {
// 		return false
// 	}
// 	for i := range a {
// 		if !matchesValue(a[i], b[i]) {
// 			return false
// 		}
// 	}
// 	return true
// }

// // Returns true if the values matches (must be json types)
// // The types of the values must match, otherwise it will always return false
// // If two map[string]interface{} are given, all elements must match.
// func matchesValue(av, bv interface{}) bool {
// 	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
// 		return false
// 	}
// 	switch at := av.(type) {
// 	case string:
// 		bt := bv.(string)
// 		if bt == at {
// 			return true
// 		}
// 	case float64:
// 		bt := bv.(float64)
// 		if bt == at {
// 			return true
// 		}
// 	case bool:
// 		bt := bv.(bool)
// 		if bt == at {
// 			return true
// 		}
// 	case nil:
// 		// Both nil, fine.
// 		return true
// 	case map[string]interface{}:
// 		bt := bv.(map[string]interface{})
// 		if len(bt) != len(at) {
// 			return false
// 		}
// 		for key := range bt {
// 			av, aOK := at[key]
// 			bv, bOK := bt[key]
// 			if aOK != bOK {
// 				return false
// 			}
// 			if !matchesValue(av, bv) {
// 				return false
// 			}
// 		}
// 		return true
// 	case []interface{}:
// 		bt := bv.([]interface{})
// 		return matchesArray(at, bt)
// 	}
// 	return false
// }

// // getDiff returns the (recursive) difference between a and b as a map[string]interface{}.
// func getDiff(a, b map[string]interface{}) (map[string]interface{}, error) {
// 	into := map[string]interface{}{}
// 	for key, bv := range b {
// 		av, ok := a[key]
// 		// value was added
// 		if !ok {
// 			into[key] = bv
// 			continue
// 		}
// 		// If types have changed, replace completely
// 		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
// 			into[key] = bv
// 			continue
// 		}
// 		// Types are the same, compare values
// 		switch at := av.(type) {
// 		case map[string]interface{}:
// 			bt := bv.(map[string]interface{})
// 			dst := make(map[string]interface{}, len(bt))
// 			dst, err := getDiff(at, bt)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if len(dst) > 0 {
// 				into[key] = dst
// 			}
// 		case string, float64, bool:
// 			if !matchesValue(av, bv) {
// 				into[key] = bv
// 			}
// 		case []interface{}:
// 			bt := bv.([]interface{})
// 			if !matchesArray(at, bt) {
// 				into[key] = bv
// 			}
// 		case nil:
// 			switch bv.(type) {
// 			case nil:
// 				// Both nil, fine.
// 			default:
// 				into[key] = bv
// 			}
// 		default:
// 			panic(fmt.Sprintf("Unknown type:%T in key %s", av, key))
// 		}
// 	}
// 	// Now add all deleted values as nil
// 	for key := range a {
// 		_, found := b[key]
// 		if !found {
// 			into[key] = nil
// 		}
// 	}
// 	return into, nil
// }
