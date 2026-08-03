package fdsnwsgo

import (
	"bytes"
	"encoding"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	component "github.com/bclswl0827/fdsnws-go/service"
)

func encodeResult(outcome component.Outcome, value any) (Response, error) {
	statusCode := outcome.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode < 100 || statusCode > 599 {
		return Response{}, fmt.Errorf("service returned invalid HTTP status %d", statusCode)
	}
	response := Response{StatusCode: statusCode, Headers: map[string]string{}}
	if result, ok := value.(Result); ok {
		value = result.Body
		response.MIMEType = result.MIMEType
		for key, headerValue := range result.Headers {
			response.Headers[key] = headerValue
		}
	} else if result, ok := value.(*Result); ok && result != nil {
		value = result.Body
		response.MIMEType = result.MIMEType
		for key, headerValue := range result.Headers {
			response.Headers[key] = headerValue
		}
	}
	if isNil(value) {
		return Response{}, fmt.Errorf("service returned an empty Result body")
	}

	if response.MIMEType == "" {
		response.MIMEType = outcome.MIMEType
		if response.MIMEType == "" && outcome.Format == "text" {
			response.MIMEType = MIMEText
		} else if response.MIMEType == "" {
			response.MIMEType = MIMEXML
		}
	}

	var err error
	switch body := value.(type) {
	case []byte:
		response.Body = append([]byte(nil), body...)
	case string:
		response.Body = []byte(body)
	case io.Reader:
		response.Body, err = io.ReadAll(body)
	case []string:
		if outcome.ListRoot != "" && outcome.ListItem != "" {
			response.Body, err = encodeStringListXML(outcome.ListRoot, outcome.ListItem, body)
		} else {
			response.Body = []byte(strings.Join(body, "\n"))
		}
	default:
		if response.MIMEType == MIMEText {
			if marshaler, ok := value.(encoding.TextMarshaler); ok {
				response.Body, err = marshaler.MarshalText()
			} else {
				response.Body = fmt.Append(nil, value)
			}
		} else if strings.HasPrefix(response.MIMEType, "application/xml") {
			response.Body, err = marshalXMLValue(outcome.XMLRoot, value)
		} else {
			return Response{}, fmt.Errorf("service result type %T requires bytes or string for MIME type %q", value, response.MIMEType)
		}
	}
	if err != nil {
		return Response{}, fmt.Errorf("encode service result: %w", err)
	}
	return response, nil
}

func encodeStringListXML(root, item string, values []string) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "  ")
	if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: root}}); err != nil {
		return nil, err
	}
	for _, value := range values {
		if err := encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: item}}); err != nil {
			return nil, err
		}
	}
	if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: root}}); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

func marshalXMLValue(root string, value any) ([]byte, error) {
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Interface || reflected.Kind() == reflect.Pointer) {
		if reflected.IsNil() {
			return nil, fmt.Errorf("cannot XML-encode a nil value")
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() != reflect.Map {
		body, err := xml.MarshalIndent(value, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(append([]byte(xml.Header), body...), '\n'), nil
	}
	if root == "" {
		root = "Result"
	}
	return marshalMapXML(root, reflected)
}

func marshalMapXML(defaultRoot string, value reflect.Value) ([]byte, error) {
	if value.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("XML map keys must be strings")
	}
	keys := sortedMapKeys(value)
	root := defaultRoot
	content := value
	if len(keys) == 1 && validXMLName(keys[0]) {
		root = keys[0]
		content = value.MapIndex(mapKey(value, keys[0]))
	}
	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "  ")
	if err := encodeXMLNode(encoder, root, content); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

func encodeXMLNode(encoder *xml.Encoder, name string, value reflect.Value) error {
	if !validXMLName(name) {
		return fmt.Errorf("invalid XML element name %q", name)
	}
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	start := xml.StartElement{Name: xml.Name{Local: name}}
	if !value.IsValid() {
		return encoder.EncodeElement("", start)
	}
	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("XML map keys must be strings")
		}
		if err := encoder.EncodeToken(start); err != nil {
			return err
		}
		for _, key := range sortedMapKeys(value) {
			if err := encodeXMLNode(encoder, key, value.MapIndex(mapKey(value, key))); err != nil {
				return err
			}
		}
		return encoder.EncodeToken(start.End())
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if err := encodeXMLNode(encoder, name, value.Index(index)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		return encoder.EncodeElement(value.Interface(), start)
	default:
		return encoder.EncodeElement(fmt.Sprint(value.Interface()), start)
	}
}

func mapKey(value reflect.Value, key string) reflect.Value {
	return reflect.ValueOf(key).Convert(value.Type().Key())
}

func sortedMapKeys(value reflect.Value) []string {
	keys := make([]string, 0, value.Len())
	iterator := value.MapRange()
	for iterator.Next() {
		keys = append(keys, iterator.Key().String())
	}
	sort.Strings(keys)
	return keys
}

func validXMLName(name string) bool {
	if name == "" {
		return false
	}
	for index, char := range name {
		if index == 0 {
			if !(char == '_' || char == ':' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z') {
				return false
			}
		} else if !(char == '_' || char == ':' || char == '-' || char == '.' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9') {
			return false
		}
	}
	return true
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
