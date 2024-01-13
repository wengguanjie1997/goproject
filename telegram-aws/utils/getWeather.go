package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GeoRespData struct {
	Code string `json:"code"`
	Data []struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	} `json:"location"`
}
type WeatherRespData struct {
	Code string `json:"code"`
	Data struct {
		Temperature string `json:"temp"`
		Weather     string `json:"text"`
		Date        string `json:"obsTime"`
	} `json:"now"`
}
type WeatherData struct {
	Name        string `json:"cityName"`
	Temperature string `json:"temperature"`
	Date        string `json:"date"`
}

// 封装http的Get请求
func httpGetUrl(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %v ", err)
	}
	return body, nil
}

// 时间格式转换
func convertTimeFormat(inputTime string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04-07:00", inputTime)
	if err != nil {
		return "", err
	}

	formattedTime := t.Format("2006-01-02 15:04")
	return formattedTime, nil
}

// 获取城市的id
func getCityId(cityName string, apiKey string) (*GeoRespData, error) {

	geoUrl := fmt.Sprintf("https://geoapi.qweather.com/v2/city/lookup?location=%s&key=%s", cityName, apiKey)
	body, err := httpGetUrl(geoUrl)
	if err != nil {
		return nil, fmt.Errorf("http get fail: %v", err)
	}
	var geoResp GeoRespData
	err = json.Unmarshal(body, &geoResp)
	if err != nil {
		return nil, fmt.Errorf("json the data failed: %v", err)
	}
	if geoResp.Code == "200" && len(geoResp.Data) > 0 {
		return &geoResp, nil
	} else if geoResp.Code == "404" {
		return nil, fmt.Errorf("get data failed 404")
	} else {
		return nil, fmt.Errorf("error get args")
	}

}

// GetCityWeather 通过城市id获取城市的天气状况和温度

func GetCityWeather(cityName string, apiKey string) (*WeatherData, error) {
	// var cityId string
	geoResp, err := getCityId(cityName, apiKey)
	if err != nil {
		fmt.Println(err)
	}
	if geoResp == nil {
		return &WeatherData{}, nil
	}
	cityId := geoResp.Data[0].Id
	weatherUrl := fmt.Sprintf("https://devapi.qweather.com/v7/weather/now?location=%s&key=%s", cityId, apiKey)
	body, err := httpGetUrl(weatherUrl)
	if err != nil {
		return nil, fmt.Errorf("http get fail: %v", err)
	}
	var weatherResp WeatherRespData
	err = json.Unmarshal(body, &weatherResp)

	if err != nil {
		return nil, fmt.Errorf("json the data failed: %v", err)
	}

	var weatherData WeatherData

	if weatherResp.Code == "200" {
		date, _ := convertTimeFormat(weatherResp.Data.Date)
		weatherData = WeatherData{
			Name:        geoResp.Data[0].Name,
			Date:        date,
			Temperature: weatherResp.Data.Temperature,
		}
		return &weatherData, nil
	} else if weatherResp.Code == "404" {
		return nil, fmt.Errorf("get data failed 404")
	} else {
		return nil, fmt.Errorf("error get args")
	}

}

//func main() {
//	cityName := "shenzhen"
//	weatherInfo, err := getCityWeather(cityName)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Printf("城市：%s\n温度：%s\n日期：%s\n", weatherInfo.Name, weatherInfo.Temperature, weatherInfo.Date)
//}
