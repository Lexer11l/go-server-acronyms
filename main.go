package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

var cacheMap map[string][]string
var dictionaryFileName = "dictionary.json"

//ParseBody method parses body of POST request
func ParseBody(r *http.Request) []string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var t []string
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	return t
}

//LoadHandler - handler for /load
func LoadHandler(w http.ResponseWriter, r *http.Request) {
	dictionary := ParseBody(r)
	UpdateCache(dictionary)
	SaveFile()
}

//EncodeWord creates code for using as key
func EncodeWord(word string) string {
	letters := make(map[string]int)
	for i := 0; i < len(word); i++ {
		letter := string(word[i])
		_, isPresent := letters[letter]
		if isPresent {
			letters[letter]++
		} else {
			letters[letter] = 1
		}
	}

	keys := make([]string, 0, len(letters))
	for k := range letters {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := ""

	for _, key := range keys {
		result += key + strconv.Itoa(letters[key])
	}
	return result
}

//UpdateCache method updates cache map with new dictionary
func UpdateCache(words []string) {
	cacheMap = make(map[string][]string)
	for _, originalWord := range words {
		word := strings.ToLower(originalWord)
		encodedWord := EncodeWord(word)
		value, ok := cacheMap[encodedWord]
		if ok {
			cacheMap[encodedWord] = append(value, word)
		} else {
			cacheMap[encodedWord] = []string{word}
		}
	}
	log.Println("map was updated", cacheMap)
}

//SaveFile method saves content of cache map to file to restore it after restart
func SaveFile() {
	file, marshalError := json.MarshalIndent(cacheMap, "", " ")
	if marshalError != nil {
		log.Println("error", marshalError)
	}
	writeFileError := ioutil.WriteFile(dictionaryFileName, file, 0644)
	if writeFileError == nil {
		log.Println("changes saved to file")
	} else {
		log.Println("error", writeFileError)
	}

}

//GetFromCache method retrieves acronyms for passed word from dictionary
func GetFromCache(word string) []string {
	return cacheMap[EncodeWord(word)]
}

//LoadDictionaryFromFile methods restores cache map from file
func LoadDictionaryFromFile() {
	cacheMap = make(map[string][]string)
	file, err := ioutil.ReadFile(dictionaryFileName)
	if err != nil {
		log.Println("error during reading file", err)
	}
	err = json.Unmarshal(file, &cacheMap)
	if err != nil {
		log.Println("unmarshalling error", err)
	} else {
		log.Println("current dictionary:", cacheMap)
	}
}

//GetHandler is a handle for /get path
func GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	wordToCheck := r.Form.Get("word")
	cachedValue := GetFromCache(wordToCheck)
	json.NewEncoder(w).Encode(cachedValue)
}

func main() {
	LoadDictionaryFromFile()
	http.HandleFunc("/get", GetHandler)
	http.HandleFunc("/load", LoadHandler)
	http.ListenAndServe(":8080", nil)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
