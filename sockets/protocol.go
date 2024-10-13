package sockets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func stringifyValue(target reflect.Value) string {
	switch kind := target.Type().Kind(); kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(target.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(target.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(target.Float(), 'E', -1, 64)
	case reflect.String:
		return target.String()
	case reflect.Slice:
		result := make([]string, 0)
		cnt := target.Len()
		for i := 0; i < cnt; i++ {
			valOf := target.Index(i)
			value := stringifyValue(valOf)
			if strings.Index(value, " ") > 0 {
				value = "\"" + value + "\""
			}
			result = append(result, value)
		}
		return "[]" + strings.Join(result, ",")
	}
	return ""
}

// https://golang.hotexamples.com/site/file?hash=0xf399f1c463dffb8c34a75bf38132651e5615db421a2dcaee3935d79655de28fb&fullName=redis-store.go&project=antongulenko/http-isolation-proxy
func parseAndSet(target reflect.Value, val string) error {
	if !target.CanSet() {
		return fmt.Errorf("cannot set %v to %v", target, val)
	}
	switch kind := target.Type().Kind(); kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			target.SetInt(intVal)
		}
		return err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intVal, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			target.SetUint(intVal)
		}
		return err
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err == nil {
			target.SetFloat(floatVal)
		}
		return err
	case reflect.String:
		target.SetString(val)
		return nil
	}
	return fmt.Errorf("field %v has type %v, cannot set to %v", target, target.Type(), val)
}

func ExtractFromInterface(obj interface{}) []string {
	val := reflect.ValueOf(obj).Elem()

	numFields := val.NumField()
	fields := make([]string, numFields)

	for i := 0; i < numFields; i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if valueField.IsValid() {
			metadataOrder := typeField.Tag.Get("fieldOrder")
			order, _ := strconv.Atoi(metadataOrder)

			f := valueField.Interface()
			val := reflect.ValueOf(f)

			finalVal := stringifyValue(val)
			finalVal = strings.Trim(finalVal, " ")

			if ((len(finalVal) > 1 && finalVal[0:2] != "[]") && strings.Index(finalVal, " ") > 0) || len(finalVal) <= 0 {
				finalVal = "\"" + finalVal + "\""
			}

			fields[order] = finalVal
		}

	}
	return fields
}

func injectIntoInterface(val interface{}, msgContent string) {
	valueR, _ := regexp.Compile(`(?:^|\s)((?:(?:(?:[\\]["])|(?:["]))(?P<val1>(?:|[^\\"])+)?(?:(?:[\\]["])|(?:["])))|(?P<val2>[^\s"]+))`)

	valueMatches := valueR.FindAllStringSubmatch(msgContent, -1)
	val1Idx := valueR.SubexpIndex("val1")
	val2Idx := valueR.SubexpIndex("val2")
	orderedValues := make([]string, 0)
	for _, match := range valueMatches {
		val1 := match[val1Idx]
		val2 := match[val2Idx]
		value := val1
		if value == "" {
			value = val2
		}

		orderedValues = append(orderedValues, value)
	}

	valOf := reflect.ValueOf(val).Elem()
	typeOf := reflect.TypeOf(val).Elem()

	for i := 0; i < valOf.NumField(); i++ {
		fieldValue := valOf.Field(i)
		fieldType := typeOf.Field(i)

		if fieldValue.IsValid() {
			metadataOrder := fieldType.Tag.Get("fieldOrder")

			order, _ := strconv.Atoi(metadataOrder)
			value := orderedValues[order]

			parseAndSet(fieldValue, value)
		}
	}
}

func determineMessageType(message []byte) (int, int, string, error) {
	if len(message) < 4 {
		return 0, 0, "", errors.New("message too short")
	}

	// Extract the first 2 bytes as the msgType in big endian
	msgType := int(binary.BigEndian.Uint16(message[:2]))

	// Extract the next 2 bytes as the content length in big endian
	contentLength := int(binary.BigEndian.Uint16(message[2:4]))

	// Verify the message length matches the content length
	if len(message[4:]) < contentLength {
		return 0, 0, "", errors.New("message length mismatch")
	}

	// Extract the content based on the content length
	msgContent := string(message[4 : 4+contentLength])

	return msgType, contentLength, msgContent, nil
}

func SerializeObject(val interface{}) string {
	fields := ExtractFromInterface(val)
	return strings.Join(fields, " ")
}

func SerializeMessage(msgType int, val interface{}) []byte {
	content := SerializeObject(val)
	contentLength := len(content)
	// Create a buffer to hold the message
	var buffer bytes.Buffer

	// Convert msgType to big endian (2 bytes)
	msgTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(msgTypeBytes, uint16(msgType))

	// Convert contentLength to big endian (2 bytes)
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(contentLength))

	// Write the header to the buffer
	buffer.Write(msgTypeBytes)
	buffer.Write(lengthBytes)

	// Write the content (the actual serialized message)
	buffer.Write([]byte(content))

	// fmt.Println("Sending: %d, %d: %s", msgType, contentLength, content)

	// Return the final byte slice
	return buffer.Bytes()
}

func DeserializeMessage[TMessage interface{}](obj TMessage, msg string) TMessage {
	injectIntoInterface(&obj, msg)
	return obj
}
