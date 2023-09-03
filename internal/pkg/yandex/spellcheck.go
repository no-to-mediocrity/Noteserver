package yandex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"noteserver/internal/pkg/models"
	"time"
)

const (
	yaURL = "https://speller.yandex.net/services/spellservice.json/checkText?text="
)

func Spellcheck(input string, timeout int) ([]models.SpellcheckData, error) {
	encodedString := url.QueryEscape(input)
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	response, err := client.Get(yaURL + encodedString)
	if err != nil {
		return []models.SpellcheckData{}, err
	}
	defer response.Body.Close()

	jsonData, _ := ioutil.ReadAll(response.Body)
	var spellcheckDataList []models.SpellcheckData
	if response.StatusCode == 200 {
		_ = json.Unmarshal([]byte(jsonData), &spellcheckDataList)
	} else {
		return []models.SpellcheckData{}, fmt.Errorf("error, status code: %v", response.StatusCode)
	}
	return spellcheckDataList, nil
}
