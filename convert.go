package main

import "bytes"

/*
Expected feature for this function:
  - Remove trailing commas
  - Replace control characters in a string with their corresponding escape sequences.
  - Fix missing-comma issue between brackets

This function must be flexible to any input and preserve the original JSON structure.
(Do not destroy ANYTHING unless invalid JSON)
*/

var controlCharacterMap = map[rune]string{
	'\n': "\\n",
	'\t': "\\t",
	'\r': "",
}

func convertToValidJSON(data string) string {
	var buffer bytes.Buffer
	var subBuffer bytes.Buffer

	isEscaped := false
	isInString := false
	trailingCommaDetection := false
	missingCommaDetection := false

	skipRuneWriting := false
	skipStringCheck := false
	skipTrailingCheck := false

	forceCloseTrailDetection := false
	commaCount := 0

	for _, chr := range data {
		skipRuneWriting = false
		skipStringCheck = false
		skipTrailingCheck = false
		if !isInString && chr == ',' {
			if trailingCommaDetection || missingCommaDetection {
				buffer.WriteString(subBuffer.String())
			}

			trailingCommaDetection = true
			skipTrailingCheck = true
			skipRuneWriting = true

			subBuffer.Reset()
		}

		if trailingCommaDetection && !skipTrailingCheck && !missingCommaDetection {
			subBuffer.WriteRune(chr)
			skipRuneWriting = true

			if chr == '}' || chr == ']' {
				trailingCommaDetection = false
				buffer.WriteString(subBuffer.String())
			} else if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == '\r' || chr == ',') {
				forceCloseTrailDetection = true
			}
		}

		if missingCommaDetection {
			if !(trailingCommaDetection && skipTrailingCheck) {
				subBuffer.WriteRune(chr)
				if chr == ',' {
					commaCount++
				}
			}

			skipRuneWriting = true
			if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == '\r' || chr == ',') {
				missingCommaDetection = false

				if !(chr == '}' || chr == ']') && (trailingCommaDetection || commaCount == 0) {
					buffer.WriteRune(',')
				}

				trailingCommaDetection = false

				buffer.WriteString(subBuffer.String())
			}
		}

		if isInString && trailingCommaDetection || forceCloseTrailDetection {
			trailingCommaDetection = false
			forceCloseTrailDetection = false

			buffer.WriteRune(',')
			buffer.WriteString(subBuffer.String())
		}

		if !trailingCommaDetection && !isInString && (chr == '}' || chr == ']') && !missingCommaDetection {
			missingCommaDetection = true

			subBuffer.Reset()
		}

		if chr == '"' && !isEscaped {
			isInString = !isInString
			skipStringCheck = isInString
		}

		if isInString && (chr != '"' || isEscaped) {
			if replacement, isReplacementExists := controlCharacterMap[chr]; isReplacementExists {
				buffer.WriteString(replacement)
				skipRuneWriting = true
			}
		} else if isInString && !skipStringCheck {
			isInString = false
		}

		if chr == '\\' {
			isEscaped = !isEscaped
		} else {
			isEscaped = false
		}

		if !skipRuneWriting {
			buffer.WriteRune(chr)
		}
	}

	return buffer.String()
}

func legacyConvertToValidJson(data string) string {
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
			if chr == '\n' {
				buf.WriteString("\\n")
			} else if chr == '\t' {
				buf.WriteString("\\t")
			} else if chr != '\r' {
				buf.WriteRune(chr)
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
	bracketCommaMissingDetection := false

	data = buf.String()

	buf.Reset()
	detectionBuffer.Reset()

	for _, chr := range data {
		if !bracketsDetection && !bracketCommaMissingDetection {
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

					if chr == ']' {
						listLayer--
						bracketCommaMissingDetection = true
					}

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

		// NOTE: This code is to fix the `]\n\t"decorations": [` issue
		if bracketCommaMissingDetection {
			detectionBuffer.WriteRune(chr)
			if listLayer == 0 {
				if chr == ',' || chr == '}' || chr == ']' {
					buf.WriteString(detectionBuffer.String())
					detectionBuffer.Reset()

					bracketCommaMissingDetection = false
				} else if !(chr == ' ' || chr == '\n' || chr == '\t' || chr == '\r') {
					buf.WriteRune(',')
					buf.WriteString(detectionBuffer.String())
					bracketCommaMissingDetection = false

					detectionBuffer.Reset()
				}
			} else {
				buf.WriteString(detectionBuffer.String())
				detectionBuffer.Reset()

				bracketCommaMissingDetection = false
			}
		}

		if !isInString {
			if chr == '[' {
				listLayer++
			} else if chr == ']' {
				listLayer--
				bracketCommaMissingDetection = true
				detectionBuffer.Reset()
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
