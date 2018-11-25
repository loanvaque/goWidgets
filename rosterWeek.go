package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"html/template"
	"os"
	"strconv"
	"time"
	"sort"
)

type DatasetConfig struct {
	FileName string `json:"fileName"`
	Template struct {
		FileName string `json:"fileName"`
		Title string `json:"title"`
	}
	Shifts []struct {
		Id int `json:"id"`
		HourStart string `json:"hourStart"`
		HourEnd string `json:"hourEnd"`
	} `json:"shifts"`
	TeamList []string `json:"teamList"`
	RosterMatrix [][]int `json:"rosterMatrix"`
}

type TemplateOut struct {
	Title string
	WeekInfo []string
	Shifts []struct {
		Id int `json:"id"`
		HourStart string `json:"hourStart"`
		HourEnd string `json:"hourEnd"`
	}
	Table0 []struct {
		Name string
		Shifts []int
	}
	Table1 []struct {
		Shift int
		Names []string
	}
}

func main() () {
	// construct default config
	config := &DatasetConfig{}
	config.FileName = "rosterWeek.json"
	// load config from file
	byteString, err := ioutil.ReadFile(config.FileName)
	if err != nil { fmt.Fprintf(os.Stderr, "unable to read config file > %v\n", err); os.Exit(1) }
	if !json.Valid(byteString) { fmt.Fprintf(os.Stderr, "config file content is not json compliant\n"); os.Exit(1) }
	err = json.Unmarshal(byteString, config)
	if err != nil { fmt.Fprintf(os.Stderr, "unable to unmarshal json content of config file > %v\n", err); os.Exit(1) }

	// parse command-line arguments
	if len(os.Args) < 3 { fmt.Fprintf(os.Stderr, "unexpected number of argument\n"); os.Exit(1) }
	year, err := strconv.Atoi(os.Args[1])
	if err != nil { fmt.Fprintf(os.Stderr, "unable to parse argument for year > %v\n", err); os.Exit(1) }
	if year < 2018 || year > 2025 { fmt.Fprintf(os.Stderr, "year out of range\n"); os.Exit(1) }
	week, err := strconv.Atoi(os.Args[2])
	if err != nil { fmt.Fprintf(os.Stderr, "unable to parse argument for week > %v\n", err); os.Exit(1) }
	if week < 1 || week > 55 { fmt.Fprintf(os.Stderr, "week out of range\n"); os.Exit(1) }

	// find date references for requested week
	dateSeeker := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	y, w := dateSeeker.ISOWeek()
	for y == year && w != week {
		dateSeeker = dateSeeker.AddDate(0, 0, 7)
		y, w = dateSeeker.ISOWeek()
	}
	weekMonday := dateSeeker
	for weekMonday.Weekday().String() != "Monday" { weekMonday = weekMonday.AddDate(0, 0, -1) }
	weekSunday := weekMonday.AddDate(0, 0, 6)

	shiftedTeamlist := []string{}
	shiftedIndex := week % len(config.TeamList)
	for len(shiftedTeamlist) < len(config.TeamList) {
		shiftedTeamlist = append(shiftedTeamlist, config.TeamList[shiftedIndex])
		shiftedIndex++
		if shiftedIndex >= len(config.TeamList) { shiftedIndex = 0 }
	}

	// populate variable for template
	templateOut := TemplateOut{}
	templateOut.Title = config.Template.Title
	templateOut.WeekInfo = []string{strconv.Itoa(year), strconv.Itoa(week), weekMonday.Format("02/01"), weekSunday.Format("02/01")}
	templateOut.Shifts = config.Shifts
	sort.Slice(templateOut.Shifts, func(i, j int) bool { return templateOut.Shifts[i].Id < templateOut.Shifts[j].Id })
	// populate table0 and prepare table1
	for index, row := range config.RosterMatrix {
		var table0Row struct{Name string; Shifts []int}
		table0Row.Name = shiftedTeamlist[index]
		table0Row.Shifts = row
		templateOut.Table0 = append(templateOut.Table0, table0Row)

		for _, cell := range row {
			found := false
			for _, shift := range templateOut.Table1 { if shift.Shift == cell { found = true; break } }
			if found == false { templateOut.Table1 = append(templateOut.Table1, struct{Shift int; Names []string}{cell, []string{},}) }
		}
	}
//	sort.Slice(templateOut.Table0, func(i, j int) bool { return templateOut.Table0[i].Name < templateOut.Table0[j].Name })
	sort.Slice(templateOut.Table1, func(i, j int) bool { return templateOut.Table1[i].Shift < templateOut.Table1[j].Shift })
	// populate table1
	for _, row0 := range templateOut.Table0 {
		for index0, cell0 := range row0.Shifts {
			for index1, row1 := range templateOut.Table1 {
				if row1.Shift == cell0 {
					for len(templateOut.Table1[index1].Names) < index0 + 1 {
						templateOut.Table1[index1].Names = append(templateOut.Table1[index1].Names, "")
					}
					if len(templateOut.Table1[index1].Names[index0]) < 1 {
						templateOut.Table1[index1].Names[index0] = row0.Name
					} else {
						templateOut.Table1[index1].Names[index0] += "\n" + row0.Name
					}
				}
			}
		}
	}

	// parse template
	tmpl, err := template.ParseFiles(config.Template.FileName)
	if err != nil { fmt.Fprintf(os.Stderr, "unable to parse the template file > %v\n", err); os.Exit(1) }

	fileHandle, err := os.OpenFile("./output.html", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0666)
	if err != nil { fmt.Fprintf(os.Stderr, "unable to write to the output file > %v\n", err); os.Exit(1) }
	defer fileHandle.Close()

	err = tmpl.Execute(fileHandle, templateOut)
	if err != nil { fmt.Fprintf(os.Stderr, "unable to execute the template > %v\n", err); os.Exit(1) }
	return
}
