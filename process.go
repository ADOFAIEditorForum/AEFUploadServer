package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var defaultFileList = []string{"level.adofai", "main.adofai"}

func detectADOFAIFile(destination string) string {
	adofaiFileName := ""
	files, err2 := os.ReadDir(destination)
	if err2 != nil {
		log.Fatal(err2)
		return ""
	}

	if len(files) == 1 && files[0].IsDir() {
		fileName := files[0].Name()
		target := filepath.Join(destination, fileName)
		return filepath.Join(fileName, detectADOFAIFile(target))
	} else {
		for _, file := range files {
			name := file.Name()

			if strings.HasSuffix(name, ".adofai") {
				if slices.Contains(defaultFileList, name) {
					adofaiFileName = name
					break
				}
				if adofaiFileName == "" && name != "backup.adofai" {
					adofaiFileName = name
				}
			}
		}

		return adofaiFileName
	}
}

var pathMap = map[rune]int{
	'R': 0,
	'p': 15,
	'J': 30,
	'E': 45,
	'T': 60,
	'o': 75,
	'U': 90,
	'q': 105,
	'G': 120,
	'Q': 135,
	'H': 150,
	'W': 165,
	'L': 180,
	'x': 195,
	'N': 210,
	'Z': 225,
	'F': 240,
	'V': 255,
	'D': 270,
	'Y': 285,
	'B': 300,
	'C': 315,
	'M': 330,
	'A': 345,
	'!': 999,
}

var vertexMap = map[rune]struct {
	int
	bool
}{
	'5': {5, false},
	'6': {5, true},
	'7': {7, false},
	'8': {7, true},
}

func getVertex(path rune) (int, bool) {
	result := vertexMap[path]
	return result.int, result.bool
}

var boolCheckSettings = []string{
	"separateCountdownTime", "seizureWarning",
	"showDefaultBGIfNoImage", "showDefaultBGTile",
	"imageSmoothing", "lockRot", "loopBG",
	"pulseOnFloor", "loopVideo", "floorIconOutlines", "stickToFloors",
	"legacyFlash", "legacyCamRelativeTo", "legacySpriteTiles",
}

var settingsDefault = map[string]interface{}{
	"version":                 13,
	"artist":                  "",
	"specialArtistType":       "None",
	"artistPermission":        "",
	"song":                    "",
	"author":                  "",
	"separateCountdownTime":   true,
	"previewImage":            "",
	"previewIcon":             "",
	"previewIconColor":        "003f52",
	"previewSongStart":        0,
	"previewSongDuration":     10,
	"seizureWarning":          false,
	"levelDesc":               "",
	"levelTags":               "",
	"artistLinks":             "",
	"speedTrialAim":           0,
	"difficulty":              1,
	"requiredMods":            []string{},
	"songFilename":            "",
	"bpm":                     100,
	"volume":                  100,
	"offset":                  0,
	"pitch":                   100,
	"hitsound":                "Kick",
	"hitsoundVolume":          100,
	"countdownTicks":          4,
	"trackColorType":          "Single",
	"trackColor":              "debb7b",
	"secondaryTrackColor":     "ffffff",
	"trackColorAnimDuration":  2,
	"trackColorPulse":         "None",
	"trackPulseLength":        10,
	"trackStyle":              "Standard",
	"trackTexture":            "",
	"trackTextureScale":       1,
	"trackGlowIntensity":      100,
	"trackAnimation":          "None",
	"beatsAhead":              3,
	"trackDisappearAnimation": "None",
	"beatsBehind":             4,
	"backgroundColor":         "000000",
	"showDefaultBGIfNoImage":  true,
	"showDefaultBGTile":       true,
	"defaultBGTileColor":      "101121",
	"defaultBGShapeType":      "Default",
	"defaultBGShapeColor":     "ffffff",
	"bgImage":                 "",
	"bgImageColor":            "ffffff",
	"parallax":                []int{100, 100},
	"bgDisplayMode":           "FitToScreen",
	"imageSmoothing":          true,
	"lockRot":                 false,
	"loopBG":                  false,
	"scalingRatio":            100,
	"relativeTo":              "Player",
	"position":                []int{0, 0},
	"rotation":                0,
	"zoom":                    100,
	"pulseOnFloor":            true,
	"startCamLowVFX":          false,
	"bgVideo":                 "",
	"loopVideo":               false,
	"vidOffset":               0,
	"floorIconOutlines":       false,
	"stickToFloors":           true,
	"planetEase":              "Linear",
	"planetEaseParts":         1,
	"planetEasePartBehavior":  "Mirror",
	"customClass":             "",
	"defaultTextColor":        "ffffff",
	"defaultTextShadowColor":  "00000050",
	"congratsText":            "",
	"perfectText":             "",
	"legacyFlash":             false,
	"legacyCamRelativeTo":     false,
	"legacySpriteTiles":       false,
}

var settingsConversion = map[string]string{
	"useLegacyFlash": "legacyFlash",
}

func process(filename string, id int64) {
	dest := fmt.Sprintf("level%d", id)

	err := unzipSource(filename, dest)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.Remove(filename)
	if err != nil {
		return
	}

	adofaiFileName := detectADOFAIFile(dest)

	println(adofaiFileName)

	data, err := os.ReadFile(filepath.Join(dest, adofaiFileName))
	trimmedBytes := bytes.Trim(data, "\xef\xbb\xbf")
	adofaiLevelStr := string(trimmedBytes)

	adofaiLevelStr = convertToValidJSON(adofaiLevelStr)
	err = os.WriteFile("log.txt", []byte(adofaiLevelStr), 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	var adofaiLevelJson map[string]interface{}
	err = json.Unmarshal([]byte(adofaiLevelStr), &adofaiLevelJson)

	if err != nil {
		log.Fatal(err)
		return
	}

	if val, ok := adofaiLevelJson["pathData"]; ok {
		var angleData []float32
		pathData := val.(string)
		lastAngle := float32(0)

		for _, path := range pathData {
			if angle, ok := pathMap[path]; ok {
				fAngle := float32(angle)
				angleData = append(angleData, fAngle)
				if angle != 999 {
					lastAngle = fAngle
				} else {
					lastAngle -= 180
				}

				// NOTE: Verify Fix
			} else {
				vertex, reverse := getVertex(path)
				vertexCalc := float32(vertex)

				relativeAngle := 180.0 - 180.0*(vertexCalc-2)/vertexCalc // TODO: Fix angleData Conversion
				if reverse {
					relativeAngle = -relativeAngle
				}

				angleData = append(angleData, lastAngle+relativeAngle)
				lastAngle += relativeAngle
			}
		}

		delete(adofaiLevelJson, "pathData")
		adofaiLevelJson["angleData"] = angleData
	}

	adofaiSettings := adofaiLevelJson["settings"].(map[string]interface{})
	for original, target := range settingsConversion {
		if _, ok := adofaiSettings[original]; ok {
			adofaiSettings[target] = adofaiSettings[original]
			delete(adofaiSettings, original)
		}
	}

	for settingName, settingValue := range settingsDefault {
		if _, ok := adofaiSettings[settingName]; !ok {
			adofaiSettings[settingName] = settingValue
		}
	}

	for _, boolCheckKey := range boolCheckSettings {
		if content, isString := adofaiSettings[boolCheckKey].(string); isString {
			switch content {
			case "Enabled":
				adofaiSettings[boolCheckKey] = true
			case "Disabled":
				adofaiSettings[boolCheckKey] = false
			default:
				println(boolCheckKey, content)
				return
			}
		}
	}

	if _, ok := adofaiLevelJson["decorations"]; !ok {
		var newActions []map[string]interface{}
		//goland:noinspection GoPreferNilSlice
		var decorations = []map[string]interface{}{}

		actions := adofaiLevelJson["actions"].([]interface{})
		for _, action := range actions {
			action := action.(map[string]interface{})
			if action["eventType"] == "AddDecoration" || action["eventType"] == "AddObject" {
				decorations = append(decorations, action)
			} else {
				newActions = append(newActions, action)
			}
		}

		adofaiLevelJson["actions"] = newActions
		adofaiLevelJson["decorations"] = decorations
	}

	jsonData, err := json.Marshal(adofaiLevelJson)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.WriteFile(filepath.Join(dest, adofaiFileName), jsonData, 0644)
	if err != nil {
		return
	}

	apiURL := "http://localhost:3677/level"
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		println(err)
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	println(string(responseBody))

	return
}

func unzipSource(filename string, destination string) error {
	cmd := exec.Command("7z", "x", filename, "-o./"+destination)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
