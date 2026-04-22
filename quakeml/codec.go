package quakeml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

func NewDocument(publicID string) *Document {
	return &Document{
		XMLNS:      Namespace,
		Parameters: EventParameters{XMLNS: BEDNamespace, PublicID: publicID},
	}
}

func Parse(data []byte) (*Document, error) {
	return Decode(bytes.NewReader(data))
}

func Decode(reader io.Reader) (*Document, error) {
	var document Document
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode QuakeML: %w", err)
	}
	if document.XMLNS == "" {
		document.XMLNS = document.XMLName.Space
	}
	if document.Parameters.XMLNS == "" {
		document.Parameters.XMLNS = document.Parameters.XMLName.Space
	}
	return &document, nil
}

func Marshal(document *Document) ([]byte, error) {
	var buffer bytes.Buffer
	if err := Encode(&buffer, document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Encode(writer io.Writer, document *Document) error {
	if document == nil {
		return fmt.Errorf("encode QuakeML: document is nil")
	}
	value := *document
	if value.XMLNS == "" {
		value.XMLNS = Namespace
	}
	if value.Parameters.XMLNS == "" {
		value.Parameters.XMLNS = BEDNamespace
	}
	if _, err := io.WriteString(writer, xml.Header); err != nil {
		return fmt.Errorf("encode QuakeML header: %w", err)
	}
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode QuakeML: %w", err)
	}
	return nil
}
