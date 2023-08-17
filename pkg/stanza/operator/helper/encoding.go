// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package helper // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Deprecated: [v0.84.0]
func NewEncodingConfig() EncodingConfig {
	return EncodingConfig{
		Encoding: "utf-8",
	}
}

// Deprecated: [v0.84.0]
type EncodingConfig struct {
	Encoding string `mapstructure:"encoding,omitempty"`
}

type Decoder struct {
	encoding     encoding.Encoding
	decoder      *encoding.Decoder
	decodeBuffer []byte
}

// NewDecoder wraps a character set encoding and creates a reusable buffer to reduce allocation.
// Decoder is not thread-safe and must not be used in multiple goroutines.
func NewDecoder(enc encoding.Encoding) *Decoder {
	return &Decoder{
		encoding:     enc,
		decoder:      enc.NewDecoder(),
		decodeBuffer: make([]byte, 1<<12),
	}
}

// Decode converts the bytes in msgBuf to UTF-8 from the configured encoding.
func (d *Decoder) Decode(msgBuf []byte) ([]byte, error) {
	for {
		d.decoder.Reset()
		nDst, _, err := d.decoder.Transform(d.decodeBuffer, msgBuf, true)
		if err == nil {
			return d.decodeBuffer[:nDst], nil
		}
		if errors.Is(err, transform.ErrShortDst) {
			d.decodeBuffer = make([]byte, len(d.decodeBuffer)*2)
			continue
		}
		return nil, fmt.Errorf("transform encoding: %w", err)
	}
}

var encodingOverrides = map[string]encoding.Encoding{
	"utf-16":   unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	"utf16":    unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	"utf-8":    unicode.UTF8,
	"utf8":     unicode.UTF8,
	"ascii":    unicode.UTF8,
	"us-ascii": unicode.UTF8,
	"nop":      encoding.Nop,
	"":         unicode.UTF8,
}

// LookupEncoding attempts to match the string name provided with a character set encoding.
func LookupEncoding(enc string) (encoding.Encoding, error) {
	if e, ok := encodingOverrides[strings.ToLower(enc)]; ok {
		return e, nil
	}
	e, err := ianaindex.IANA.Encoding(enc)
	if err != nil {
		return nil, fmt.Errorf("unsupported encoding '%s'", enc)
	}
	if e == nil {
		return nil, fmt.Errorf("no charmap defined for encoding '%s'", enc)
	}
	return e, nil
}

func IsNop(enc string) bool {
	e, err := LookupEncoding(enc)
	if err != nil {
		return false
	}
	return e == encoding.Nop
}
