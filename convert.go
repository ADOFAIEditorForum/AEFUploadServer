package main

import (
	"bytes"
)

func convertToValidJSON(data string) string {
	var buf bytes.Buffer

	isInString := false
	isEscapedChar := false
	isCommaDetectionEnabled := false
	var detectionBuffer bytes.Buffer

	for _, chr := range data {
		if chr == '"' && !isEscapedChar {
			isInString = !isInString
			if isCommaDetectionEnabled {
				isCommaDetectionEnabled = false

				buf.WriteRune(',')
				buf.WriteString(detectionBuffer.String())
			}

			//buf.WriteRune('"')
		}

		if chr == ',' && !isInString {
			isCommaDetectionEnabled = true
			detectionBuffer.Reset()
			isEscapedChar = false
			continue
		}

		if isInString {
			if chr != '\n' {
				buf.WriteRune(chr)
			} else {
				buf.WriteString("\\n")
			}
		} else if isCommaDetectionEnabled {
			detectionBuffer.WriteRune(chr)
			if chr == '}' || chr == ']' {
				isCommaDetectionEnabled = false
				buf.WriteString(detectionBuffer.String())
			} else if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == '\r' || chr == ',') {
				isCommaDetectionEnabled = false
				buf.WriteRune(',')
				buf.WriteString(detectionBuffer.String())
			}
		} else {
			buf.WriteRune(chr)
		}

		if chr == '\\' {
			isEscapedChar = true
		} else if isEscapedChar {
			isEscapedChar = false
		}
	}

	return buf.String()
}
