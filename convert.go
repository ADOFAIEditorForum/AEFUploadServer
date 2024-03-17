package main

import (
	"bytes"
)

func convertToValidJSON(data string) string {
	var buf bytes.Buffer

	isInString := false
	isEscapedChar := false
	isCommaDetectionEnabled := false

	listLayer := 0
	bracketsDetection := false
	listLineSeparation := false

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
			isEscapedChar = !isEscapedChar
		} else if isEscapedChar {
			isEscapedChar = false
		}
	}

	isInString = false
	isEscapedChar = false

	data = buf.String()

	buf.Reset()
	detectionBuffer.Reset()

	for _, chr := range data {
		if !bracketsDetection {
			buf.WriteRune(chr)
		}

		if !isEscapedChar && chr == '"' {
			isInString = !isInString
		}

		if listLayer > 0 {
			if bracketsDetection {
				detectionBuffer.WriteRune(chr)

				if chr == ',' || chr == ']' {
					bracketsDetection = false
					buf.WriteString(detectionBuffer.String())
					detectionBuffer.Reset()

					continue
				}

				if chr == '\n' {
					listLineSeparation = true
				}

				if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == '\r') {
					if chr == '{' && listLineSeparation {
						buf.WriteRune(',')
						buf.WriteString(detectionBuffer.String())

						detectionBuffer.Reset()
					} else {
						buf.WriteString(detectionBuffer.String())
						detectionBuffer.Reset()
					}
				}
			}

			if chr == '}' {
				listLineSeparation = false
				bracketsDetection = true
			}
		} else {
			listLineSeparation = false
			bracketsDetection = false
		}

		if !isInString {
			if chr == '[' {
				listLayer++
			} else if chr == ']' {
				listLayer--
			}
		}

		if chr == '\\' {
			isEscapedChar = !isEscapedChar
		} else if isEscapedChar {
			isEscapedChar = false
		}
	}

	return buf.String()
}
