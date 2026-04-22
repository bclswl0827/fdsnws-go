package quakeml

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshalParseAndStreamCodec(t *testing.T) {
	document := NewDocument("smi:local/events")
	document.Parameters.Events = []Event{{
		PublicID: "smi:local/event-1", PreferredOriginID: "smi:local/origin-1",
		Origins: []Origin{{
			PublicID: "smi:local/origin-1",
			Time:     TimeQuantity{Value: "2024-01-02T03:04:05Z"},
			Latitude: RealQuantity{Value: 1}, Longitude: RealQuantity{Value: 2}, Depth: RealQuantity{Value: 3000},
		}},
	}}

	data, err := Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<quakeml xmlns="`+Namespace+`"`) {
		t.Fatalf("xml = %s", data)
	}
	decoded, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.XMLNS != Namespace || decoded.Parameters.XMLNS != BEDNamespace {
		t.Fatalf("namespaces = %q, %q", decoded.XMLNS, decoded.Parameters.XMLNS)
	}
	if decoded.Parameters.Events[0].Origins[0].Depth.Value != 3000 {
		t.Fatalf("decoded = %#v", decoded)
	}

	var buffer bytes.Buffer
	if err := Encode(&buffer, decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(&buffer); err != nil {
		t.Fatal(err)
	}
}

func TestParsePrefixedQuakeML(t *testing.T) {
	data := []byte(`<q:quakeml xmlns:q="` + Namespace + `"><eventParameters xmlns="` + BEDNamespace + `" publicID="smi:local/events"><event publicID="smi:local/event-1"/></eventParameters></q:quakeml>`)
	document, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if document.XMLNS != Namespace || document.Parameters.XMLNS != BEDNamespace || document.Parameters.Events[0].PublicID != "smi:local/event-1" {
		t.Fatalf("document = %#v", document)
	}
}
