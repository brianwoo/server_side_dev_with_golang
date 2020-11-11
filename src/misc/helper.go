package misc

import (
	"encoding/json"
	"os"
	"strconv"
)

const EmptyJsonString string = "{}"

func GetJsonFromJsonObjs(obj interface{}) ([]byte, error) {

	jsonBytes, err := json.Marshal(&obj)
	return jsonBytes, err
}

func GetInt64FromString(numberStr string) (int64, error) {

	return strconv.ParseInt(numberStr, 10, 64)
}

func GetEmptyJsonByteArray() []byte {
	return []byte(EmptyJsonString)
}

func GetConfigFilePath() string {
	configFilePath := os.Getenv("CONFUSION_CONFIG_PATH")
	if configFilePath == "" {
		panic("CONFUSION_CONFIG_PATH is not set!")

	} else {
		info, err := os.Stat(configFilePath)
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			panic(configFilePath + " is a directory!")
		}
		return configFilePath
	}
}
