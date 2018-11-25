package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"net/url"
)

type DatasetConfig struct {
	FileName string
	Title string `json:"title"`
	Openweathermap struct {
		ApiUrl string `json:"apiUrl"`
		ApiKey string `json:"apiKey"`
		CityId string `json:"cityId"`
	} `json:"openweathermap"`
	Highcharts struct {
		ApiUrl string `json:"apiUrl"`
	} `json:"highcharts"`
}

type OpenweathermapJsonIn struct {
	List []struct {
		Date string `json:"dt_txt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Cloud struct {
			Cover float64 `json:"all"`
		} `json:"clouds"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Rain struct {
			Precip float64 `json:"3h"`
		} `json:"rain"`
	} `json:"list"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

type HighchartsJsonOut struct {
	Chart struct {
		Type string `json:"type"`
	} `json:"chart"`
	Title struct {
		Text string `json:"text"`
	} `json:"title"`
	Subtitle struct {
		Text string `json:"text"`
	} `json:"subtitle"`
	Yaxis []HighchartsYaxis `json:"yAxis"`
	Xaxis struct {
		Categories []string `json:"categories"`
		GridLineWidth int `json:"gridLineWidth"`
		Opposite bool `json:"opposite"`
	} `json:"xAxis"`
	PlotOptions struct {
		Spline struct {
			DataLabels struct {
				Enabled bool `json:"enabled"`
			} `json:"dataLabels"`
		} `json:"spline"`
		Series struct {
			Marker struct {
				Symbol string `json:"symbol"`
			} `json:"marker"`
		} `json:"series"`
	} `json:"plotOptions"`
	Series []HighchartsSerie `json:"series"`
}

type HighchartsYaxis struct {
	Min int `json:"min"`
	Max int `json:"max"`
	Visible bool `json:"visible"`
}

type HighchartsSerie struct {
	Name string `json:"name"`
	Yaxis int `json:"yAxis"`
	Color string `json:"color"`
	DataLabels HighchartsLabels `json:"dataLabels"`
	Data []int64 `json:"data"`
}

type HighchartsLabels struct {
	Format string `json:"format"`
	Style struct {
		Color string `json:"color"`
	} `json:"style"`
}

func main() () {
	// construct default config
	config := &DatasetConfig{}
	config.FileName = "weatherForecast.json"
	// load config from file
	byteString, err := ioutil.ReadFile(config.FileName)
	if err != nil {
		fmt.Printf("unable to read config file > %v\n", err)
		return
	}
	if !json.Valid(byteString) {
		fmt.Printf("config file content is not json compliant\n")
		return
	}
	err = json.Unmarshal(byteString, config)
	if err != nil {
		fmt.Printf("unable to unmarshal json content of config file > %v\n", err)
		return
	}
	// query openweathermap api
	httpResponse, err := http.Get(config.Openweathermap.ApiUrl + "?APPID=" + config.Openweathermap.ApiKey + "&id=" + config.Openweathermap.CityId + "&units=metric")
	if err != nil {
		fmt.Printf("unable to query openweathermap api > %v\n", err)
		return
	}
	defer httpResponse.Body.Close()
	// parse to json
	jsonBytesIn, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		fmt.Printf("unable to read response content > %v\n", err)
		return
	}
	jsonIn := OpenweathermapJsonIn{}
	err = json.Unmarshal(jsonBytesIn, &jsonIn)
	if err != nil {
		fmt.Printf("unable to unmarshal json payload > %v\n", err)
		return
	}
	// convert to highcharts
	highchartsJsonOut := &HighchartsJsonOut{}
	highchartsJsonOut.Chart.Type = "spline"
	highchartsJsonOut.Title.Text = config.Title
	highchartsJsonOut.Subtitle.Text = jsonIn.City.Name + ", " + jsonIn.List[0].Date[0:10]
	highchartsJsonOut.PlotOptions.Spline.DataLabels.Enabled = true
	highchartsJsonOut.PlotOptions.Series.Marker.Symbol = "circle"

	yaxis0 := HighchartsYaxis{}
	yaxis0.Min = 0
	yaxis0.Max = 30
	yaxis0.Visible = false
	highchartsJsonOut.Yaxis = append(highchartsJsonOut.Yaxis, yaxis0)

	yaxis1 := HighchartsYaxis{}
	yaxis1.Min = 0
	yaxis1.Max = 90
	yaxis1.Visible = false
	highchartsJsonOut.Yaxis = append(highchartsJsonOut.Yaxis, yaxis1)

	yaxis2 := HighchartsYaxis{}
	yaxis2.Min = 0
	yaxis2.Max = 20
	yaxis2.Visible = false
	highchartsJsonOut.Yaxis = append(highchartsJsonOut.Yaxis, yaxis2)

	yaxis3 := HighchartsYaxis{}
	yaxis3.Min = 0
	yaxis3.Max = 100
	yaxis3.Visible = false
	highchartsJsonOut.Yaxis = append(highchartsJsonOut.Yaxis, yaxis3)

	// https://www.rapidtables.com/web/color/RGB_Color.html
	serie0 := HighchartsSerie{"Temperature", 0, " #00CC00", HighchartsLabels{}, []int64{}}
	serie0.DataLabels.Style.Color = "#00CC00"
	serie0.DataLabels.Format = "{y}°C"
	serie1 := HighchartsSerie{"Rainfall", 1, "#0066CC", HighchartsLabels{}, []int64{}}
	serie1.DataLabels.Style.Color = "#0066CC"
	serie1.DataLabels.Format = "{y}mm"
	serie2 := HighchartsSerie{"Wind speed", 2, "#CC0000", HighchartsLabels{}, []int64{}}
	serie2.DataLabels.Style.Color = "#CC0000"
	serie2.DataLabels.Format = "{y}m/s"
	serie3 := HighchartsSerie{"Cloud cover", 3, "#6600CC", HighchartsLabels{}, []int64{}}
	serie3.DataLabels.Style.Color = "#6600CC"
	serie3.DataLabels.Format = "{y}%"
	for index := 0; index < 8; index++ {
		highchartsJsonOut.Xaxis.Categories = append(highchartsJsonOut.Xaxis.Categories, jsonIn.List[index].Date[11:16])
		highchartsJsonOut.Xaxis.GridLineWidth = 1
		highchartsJsonOut.Xaxis.Opposite = true
		serie0.Data = append(serie0.Data, int64(jsonIn.List[index].Main.Temp))
		serie1.Data = append(serie1.Data, int64(jsonIn.List[index].Rain.Precip))
		serie2.Data = append(serie2.Data, int64(jsonIn.List[index].Wind.Speed))
		serie3.Data = append(serie3.Data, int64(jsonIn.List[index].Cloud.Cover))
	}
	highchartsJsonOut.Series = append(highchartsJsonOut.Series, serie0)
	highchartsJsonOut.Series = append(highchartsJsonOut.Series, serie1)
	highchartsJsonOut.Series = append(highchartsJsonOut.Series, serie2)
	highchartsJsonOut.Series = append(highchartsJsonOut.Series, serie3)
	// prepare payload
	bytesOut, err := json.Marshal(&highchartsJsonOut)
	if err != nil {
		fmt.Printf("unable to marshal json payload > %v\n", err)
		return
	}
	// query highcharts
//	fmt.Printf("%v\n", string(bytesOut))
	stringEscaped := "async=true&type=png&options=" + url.QueryEscape(string(bytesOut))
	httpResponse, err = http.Post(config.Highcharts.ApiUrl, "application/x-www-form-urlencoded", bytes.NewReader([]byte(stringEscaped)))
	if err != nil {
		fmt.Printf("unable to query highcharts api > %v\n", err)
		return
	}
	defer httpResponse.Body.Close()
	// dump response
	bytesIn, err := ioutil.ReadAll(httpResponse.Body)
	fmt.Printf("%v%v\n", config.Highcharts.ApiUrl, string(bytesIn))

	return
}
