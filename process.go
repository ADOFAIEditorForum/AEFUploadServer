package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

var priority_high = []string{"level.adofai", "main.adofai"}
var priority_low = []string{"backup.adofai"}

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
		result := detectADOFAIFile(target)
		if result != "" {
		    return filepath.Join(fileName, result)
		}
		
		return ""
	} else {
		lowp_exists := ""
		for _, file := range files {
			name := file.Name()

			if !file.IsDir() && strings.HasSuffix(name, ".adofai") {
				if slices.Contains(priority_high, name) {
					adofaiFileName = name
					break
				}

				if !slices.Contains(priority_low, name) {
					if adofaiFileName == "" {
						adofaiFileName = name
					}
				} else {
					for _, target := range priority_low {
						if target == lowp_exists {
							break
						}

						if target == name {
							lowp_exists = name
							break
						}
					}
				}
			}
		}

		if adofaiFileName == "" {
			if lowp_exists != "" {
				adofaiFileName = lowp_exists
			} else {
				for _, file := range files {
					name := file.Name()
					if file.IsDir() {
						target := filepath.Join(destination, name)
						result := detectADOFAIFile(target)

						if result != "" {
							adofaiFileName = filepath.Join(name, result)
							break
						}
					}
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

var trackColorSettings = []string{
	"trackColor", "secondaryTrackColor",
}

var settingsConversion = map[string]string{
	"useLegacyFlash": "legacyFlash",
}

func calcNewAlpha(alpha int64) int64 {
	// return int64(math.Round(float64(255-alpha) * 0.4))
	floatAlpha := float64(alpha)
	if floatAlpha <= 80.174 {
		return int64(math.Round((1 - math.Pow((255-floatAlpha)/255, 6)) * 255 / 1.52240007032))
	}
	return int64(math.Round(floatAlpha + (255-floatAlpha)*0.4))
}

func removeDest(dest string) {
	err := os.RemoveAll(dest)
	if err != nil {
		log.Println(err)
	}
}

//goland:noinspection GoPreferNilSlice
func process(filename string, id int64, downloadID chan uint64) {
	dest := fmt.Sprintf("level%d", id)
	println(filename)

	err := unzipSource(filename, dest)
	if err != nil {
		log.Println(err)
		err := os.Remove(filename)
		if err != nil {
			log.Println(err)
		}

		return
	}

	err = os.Remove(filename)
	if err != nil {
		return
	}

	adofaiFileName := detectADOFAIFile(dest)
	if adofaiFileName == "" { // There is no ADOFAI file
		removeDest(dest)
		return
	}

	log.Println(adofaiFileName)

	data, err := os.ReadFile(filepath.Join(dest, adofaiFileName))
	trimmedBytes := bytes.Trim(data, "\xef\xbb\xbf")
	adofaiLevelStr := string(trimmedBytes)

	adofaiLevelStr = convertToValidJSON(adofaiLevelStr)
	err = os.WriteFile("log.txt", []byte(adofaiLevelStr), 0644)
	if err != nil {
		log.Println(err)
		removeDest(dest)
		return
	}

	var adofaiLevelJson map[string]interface{}
	err = json.Unmarshal([]byte(adofaiLevelStr), &adofaiLevelJson)

	if err != nil {
		log.Println(err)
		removeDest(dest)
		return
	}

	checkForTileAlpha := false
	if val, ok := adofaiLevelJson["pathData"]; ok {
		var angleData []float32
		pathData := val.(string)
		lastAngle := float32(0)

		checkForTileAlpha = true

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

				relativeAngle := 180.0 - 180.0*(vertexCalc-2)/vertexCalc
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

	if checkForTileAlpha {
		for _, colorCheckKey := range trackColorSettings {
			if content, isString := adofaiSettings[colorCheckKey].(string); isString {
				println("Detected: " + colorCheckKey + " = " + content)
				if len(content) == 8 {
					alphaColor, err := strconv.ParseInt(content[6:8], 16, 32)
					if err != nil {
						continue
					}

					alphaColor = calcNewAlpha(alphaColor)
					adofaiSettings[colorCheckKey] = fmt.Sprintf("%s%02x", content[:6], alphaColor)
				}
			}
		}
	}

	updateToDecorations := false
	if _, ok := adofaiLevelJson["decorations"]; !ok {
		updateToDecorations = true
	}

	// ColorTrack
	var newActions = []map[string]interface{}{}
	var decorations = []map[string]interface{}{}

	actions := adofaiLevelJson["actions"].([]interface{})
	for _, action := range actions {
		isAction := true
		action := action.(map[string]interface{})
		if updateToDecorations && (action["eventType"] == "AddDecoration" || action["eventType"] == "AddObject") {
			decorations = append(decorations, action)
			isAction = false
		}

		if isAction && checkForTileAlpha && (action["eventType"] == "ColorTrack" || action["eventType"] == "RecolorTrack") {
			for _, colorCheckKey := range trackColorSettings {
				if content, isString := action[colorCheckKey].(string); isString {
					if len(content) == 8 {
						alphaColor, err := strconv.ParseInt(content[6:8], 16, 32)
						if err != nil {
							return
						}

						alphaColor = calcNewAlpha(alphaColor)
						action[colorCheckKey] = fmt.Sprintf("%s%02x", content[:6], alphaColor)
					}
				}
			}
		}

		if isAction {
			newActions = append(newActions, action)
		}
	}

	adofaiLevelJson["actions"] = newActions
	if updateToDecorations {
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
		log.Println(err)
		return
	}

	var uploadInfo map[string]interface{}
	responseBody, err := io.ReadAll(response.Body)
	err = json.Unmarshal(responseBody, &uploadInfo)
	if err != nil {
		return
	}

	println(string(responseBody))
	dID, err := strconv.ParseUint(uploadInfo["id"].(string), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	downloadID <- dID
	directory := filepath.Join(dest, filepath.Dir(adofaiFileName))

	go uploadAll(fmt.Sprintf("http://localhost:3677/upload/%s", uploadInfo["uploadID"]), directory, "")

	return
}

func get7zExec() string {
	if runtime.GOOS == "windows" {
		return "7z"
	}

	return "7zz"
}

func unzipSource(filename string, destination string) error {
	execFile := get7zExec()

	cmd := exec.Command(execFile, "x", filename, "-o./"+destination)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
