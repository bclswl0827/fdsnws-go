package stationxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

func NewDocument() *Document {
	return &Document{XMLNS: Namespace, SchemaVersion: DefaultSchemaVersion}
}

func Parse(data []byte) (*Document, error) {
	return Decode(bytes.NewReader(data))
}

func Decode(reader io.Reader) (*Document, error) {
	var document Document
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode StationXML: %w", err)
	}
	if document.XMLNS == "" {
		document.XMLNS = document.XMLName.Space
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
		return fmt.Errorf("encode StationXML: document is nil")
	}
	value := *document
	if value.XMLNS == "" {
		value.XMLNS = Namespace
	}
	if value.SchemaVersion == "" {
		value.SchemaVersion = DefaultSchemaVersion
	}
	if _, err := io.WriteString(writer, xml.Header); err != nil {
		return fmt.Errorf("encode StationXML header: %w", err)
	}
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode StationXML: %w", err)
	}
	return nil
}
