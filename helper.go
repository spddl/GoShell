package main

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strconv"
	"syscall"
)

func UTF16PtrFromString(s string) *uint16 {
	ret, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		log.Println(err)
	}
	return ret
}

func StringToInt(value string) int {
	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func StringToUint32(value string) uint32 {
	i, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return uint32(i)
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

// func fileExists(filename string) bool {
// 	_, err := os.Stat(filename)
// 	if os.IsNotExist(err) {
// 		return false
// 	}
// 	return true
// }
