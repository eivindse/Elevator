package dataStore

import (
	"bytes"
	"../config"
	"io/ioutil"
	"strings"
	"strconv"
	"fmt"
	"os"
)
func toOneZero(lister []bool) string{
	var buffer bytes.Buffer
	for i := range lister{
		if lister[i] == true{
			buffer.WriteString("1")
		}else{
			buffer.WriteString("0")
		}
	}

	return buffer.String()
}
func DoesFileExist(filename string) bool{
	if _, err := os.Stat(filename); err != nil{
		if os.IsNotExist(err){
			return false
		}
	}
	return true
}
func CreateFile(filename string){
	save := []byte("")
	err := ioutil.WriteFile(filename, save, 0644)
	if err != nil {
			panic(err)
	}
}
func SaveList(lister []int, id int){
	var list []bool
	for k := 0; k < config.NumFloors; k++{
		list = append(list, false)
	}
	for i := range lister{
		list[lister[i]] = true
	}
	filename := fmt.Sprintf("dataStore/%s", strconv.Itoa(id))
	str := toOneZero(list)
	save := []byte(str)
	err := ioutil.WriteFile(filename, save, 0644)
	if err != nil {
			panic(err)
	}
}
func LoadList(id int) []int{
	var returnList []int
	filename := fmt.Sprintf("dataStore/%s", strconv.Itoa(id))
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
			panic(err)
	}
	str := string(dat)
	newArr:= strings.Split(str, "")
	lister := make([]bool, len(newArr))
	for i := range newArr {
		lister[i] = (newArr[i] == "1")
	}
	for i := range lister{
		if lister[i] == true{
			returnList = append(returnList, i)
		}
	}
	return returnList
}