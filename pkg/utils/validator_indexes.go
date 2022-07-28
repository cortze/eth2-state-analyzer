package utils

import (
	"encoding/json"
	"io/ioutil"
)

func GetValIndexesFromJson(filePath string) ([]uint64, error) {

	var validatorIndex []uint64
	// open file and read all the indexes
	fbytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return validatorIndex, err
	}
	err = json.Unmarshal(fbytes, &validatorIndex)

	log.Infof("Readed %d validators", len(validatorIndex))

	return validatorIndex, nil

}

// func ArrayIntersection()
