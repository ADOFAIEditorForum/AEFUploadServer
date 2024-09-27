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
	'\v': "\\u000b",
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
				commaCount = 0

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

		if !isEscaped && chr == '"' {
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
