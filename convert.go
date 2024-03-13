package main

import "bytes"

func convertToValidJSON(data string) string {
	var buf bytes.Buffer

	isInString := false
	isEscapedChar := false
	isCommaDetectionEnabled := false
	var detectionBuffer bytes.Buffer

	for _, chr := range data {
		if !(isInString || isCommaDetectionEnabled) {
			buf.WriteRune(chr)
			continue
		}

		if chr == '"' && !isEscapedChar {
			isInString = !isInString
			if isCommaDetectionEnabled {
				isInString = !isInString
				isCommaDetectionEnabled = false

				buf.WriteRune(',')
				buf.WriteString(detectionBuffer.String())
			}

			buf.WriteRune('"')
		}

		if chr == ',' && !isInString {
			isCommaDetectionEnabled = true
			detectionBuffer.Reset()
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
			if chr == '}' {
				isCommaDetectionEnabled = false
				buf.WriteString(detectionBuffer.String())
			} else if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == ',') {
				isCommaDetectionEnabled = false
				buf.WriteRune(',')
				buf.WriteString(detectionBuffer.String())
			}
		} else {

		}

		if chr == '\\' {
			isEscapedChar = true
		}
	}

	return buf.String()
}
