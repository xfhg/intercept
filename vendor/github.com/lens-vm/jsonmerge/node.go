/*
 * Copyright (c) 2014, Evan Phoenix
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

import (
	"errors"

	"github.com/valyala/fastjson"
)

// This file is copied from
// https://github.com/evanphx/json-patch/blob/07737475986b62bc0ac8681f972a3c4ff771cef6/patch.go
// and modified to use github.com/valyala/fastjson instead of encoding/json

const (
	eRaw = iota
	eDoc
	eAry
)

type lazyNode struct {
	raw   []byte
	doc   *fastjson.Value
	ary   []*fastjson.Value
	which int
}

var (
	ErrTestFailed   = errors.New("test failed")
	ErrMissing      = errors.New("missing value")
	ErrUnknownType  = errors.New("unknown object type")
	ErrInvalid      = errors.New("invalid state detected")
	ErrInvalidIndex = errors.New("invalid index referenced")
)

type partialDoc *fastjson.Value
type partialArray []*fastjson.Value

func newLazyNode(raw []byte) *lazyNode {
	return &lazyNode{raw: raw, doc: nil, ary: nil, which: eRaw}
}

// func (n *lazyNode) MarshalJSON() ([]byte, error) {
// 	switch n.which {
// 	case eRaw:
// 		return json.Marshal(n.raw)
// 	case eDoc:
// 		return json.Marshal(n.doc)
// 	case eAry:
// 		return json.Marshal(n.ary)
// 	default:
// 		return nil, ErrUnknownType
// 	}
// }

// func (n *lazyNode) UnmarshalJSON(data []byte) error {
// 	dest := make(json.RawMessage, len(data))
// 	copy(dest, data)
// 	n.raw = &dest
// 	n.which = eRaw
// 	return nil
// }

func (n *lazyNode) intoDoc() (*fastjson.Value, error) {
	if n.which == eDoc {
		return n.doc, nil
	}

	if n.raw == nil {
		return nil, ErrInvalid
	}

	val, err := fastjson.ParseBytes(n.raw)
	if err != nil {
		return nil, err
	}

	n.doc = partialDoc(val)
	n.which = eDoc
	return n.doc, nil
}

func (n *lazyNode) intoAry() (*[]*fastjson.Value, error) {
	if n.which == eAry {
		return &n.ary, nil
	}

	if n.raw == nil {
		return nil, ErrInvalid
	}

	arr, err := n.doc.Array()
	n.ary = arr

	if err != nil {
		return nil, err
	}

	n.which = eAry
	return &n.ary, nil
}

// func (n *lazyNode) compact() []byte {
// 	buf := &bytes.Buffer{}

// 	if n.raw == nil {
// 		return nil
// 	}

// 	err := json.Compact(buf, *n.raw)

// 	if err != nil {
// 		return *n.raw
// 	}

// 	return buf.Bytes()
// }

// func (n *lazyNode) tryDoc() bool {
// 	if n.raw == nil {
// 		return false
// 	}

// 	err := json.Unmarshal(*n.raw, &n.doc)

// 	if err != nil {
// 		return false
// 	}

// 	n.which = eDoc
// 	return true
// }

// func (n *lazyNode) tryAry() bool {
// 	if n.raw == nil {
// 		return false
// 	}

// 	err := json.Unmarshal(*n.raw, &n.ary)

// 	if err != nil {
// 		return false
// 	}

// 	n.which = eAry
// 	return true
// }

// func (n *lazyNode) equal(o *lazyNode) bool {
// 	if n.which == eRaw {
// 		if !n.tryDoc() && !n.tryAry() {
// 			if o.which != eRaw {
// 				return false
// 			}

// 			return bytes.Equal(n.compact(), o.compact())
// 		}
// 	}

// 	if n.which == eDoc {
// 		if o.which == eRaw {
// 			if !o.tryDoc() {
// 				return false
// 			}
// 		}

// 		if o.which != eDoc {
// 			return false
// 		}

// 		if len(n.doc) != len(o.doc) {
// 			return false
// 		}

// 		for k, v := range n.doc {
// 			ov, ok := o.doc[k]

// 			if !ok {
// 				return false
// 			}

// 			if (v == nil) != (ov == nil) {
// 				return false
// 			}

// 			if v == nil && ov == nil {
// 				continue
// 			}

// 			if !v.equal(ov) {
// 				return false
// 			}
// 		}

// 		return true
// 	}

// 	if o.which != eAry && !o.tryAry() {
// 		return false
// 	}

// 	if len(n.ary) != len(o.ary) {
// 		return false
// 	}

// 	for idx, val := range n.ary {
// 		if !val.equal(o.ary[idx]) {
// 			return false
// 		}
// 	}

// 	return true
// }
