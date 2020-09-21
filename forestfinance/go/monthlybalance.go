package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Model for the configuration input file.
type config struct {
	Input    file    `json:"input"`
	Output   file    `json:"output"`
	Columns  columns `json:"columns"`
	Labels   []label `json:"labels"`
	StopDate string  `json:"stopDate"`
}

// The path and name of the input and output files.
type file struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// Describes which entries in the input file contain what data.
type columns struct {
	Description  int                `json:"description"`
	Debit        int                `json:"debit"`
	Credit       int                `json:"credit"`
	Date         int                `json:"date"`
	FixedAmounts map[string]float64 `json:"fixedAmounts"`
}

// The column header for the output file and the regex to match against the transaction description for that column.
type label struct {
	Label string
	Regex string
}

// Transaction details from the input file.
type transaction struct {
	Amount      float64
	Date        time.Time
	Description string
}

// Transactions summarized by date according to their label.
type summary struct {
	Label  string
	Amount map[time.Time]float64
}

func newTransaction(debit float64, credit float64, date time.Time, desc string) transaction {
	return transaction{credit - debit, date, desc}
}

func newSummary(label string) summary {
	return summary{label, make(map[time.Time]float64)}
}

func filter(transactions []transaction, regex string, f func(transaction, string) bool) []transaction {
	var result []transaction
	for _, transaction := range transactions {
		if f(transaction, regex) {
			result = append(result, transaction)
		}
	}
	return result
}

var configuration config
var transactions []transaction
var summarized []summary

func main() {
	defer func() {
		if oof := recover(); oof != nil {
			fmt.Printf("%+v", oof)
		}
	}()

	var err error

	cfgFile := flag.String("config", "../../config.json", "path to configuration file")
	flag.Parse()

	configuration, err = loadConfig(*cfgFile)
	errorCheck(err)
	transactions, err = loadTransactions()
	errorCheck(err)
	summarized, err = summarize(transactions)
	errorCheck(err)
	err = export()
	errorCheck(err)
}

// Load the configuration file.
func loadConfig(cfgFile string) (config config, err error) {
	file, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return
	}

	json.Unmarshal([]byte(file), &config)

	return
}

// Load the transaction file.
func loadTransactions() (transactions []transaction, err error) {
	input, err := os.Open(configuration.Input.Path + configuration.Input.Name)
	defer input.Close()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		debit, _ := strconv.ParseFloat(line[configuration.Columns.Debit], 64)
		credit, _ := strconv.ParseFloat(line[configuration.Columns.Credit], 64)
		date, _ := time.Parse("1/2/2006", line[configuration.Columns.Date])
		// Normalized date to group by month.
		ndate := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

		var regexval string
		// Some transaction don't have a searchable description (e.g., Check 1072) and instead are defined as being divisible by a fixed recurring amount.
		for label, amount := range configuration.Columns.FixedAmounts {
			if credit != 0 && math.Mod(credit, amount) == 0 || debit != 0 && math.Mod(debit, amount) == 0 {
				regexval = "_" + label
			}
		}

		transactions = append(transactions, newTransaction(debit, credit, ndate, line[configuration.Columns.Description]+regexval))
	}

	return
}

// Summarize transactions for each label, grouped by month (date normalized to the first of the month).
func summarize(transactions []transaction) (summary []summary, err error) {
	result := make(map[string][]transaction)

	for _, transaction := range transactions {
		var match bool
		for _, label := range configuration.Labels {
			match, _ = regexp.MatchString(label.Regex, transaction.Description)
			if match {
				result[label.Label] = append(result[label.Label], transaction)
				break
			}
		}

		if !match {
			result["Uncaptured"] = append(result["Uncaptured"], transaction)
		}
		result["Total"] = append(result["Total"], transaction)
	}

	for _, label := range append(configuration.Labels, label{"Uncaptured", ""}, label{"Total", ""}) {
		sum := newSummary(label.Label)
		for _, trans := range result[label.Label] {
			sum.Amount[trans.Date] += trans.Amount
		}

		summary = append(summary, sum)
	}

	return
}

// Export the summarized results to the output file.
func export() (err error) {
	file, err := os.Create(configuration.Output.Path + configuration.Output.Name)
	defer file.Close()
	if err != nil {
		return
	}

	header := ","
	for _, label := range summarized {
		header += label.Label + ","
	}
	file.WriteString(header + "\r")

	stop, err := time.Parse("2006-01-02", configuration.StopDate)
	start := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	if err != nil {
		return
	}

	for date := start; date.After(stop); date = date.AddDate(0, -1, 0) {
		row := date.Format("2006-01-02") + ","
		for _, amnt := range summarized {
			row += strconv.FormatFloat(amnt.Amount[date], 'f', 2, 64) + ","
		}
		file.WriteString(row + "\n")
	}

	return
}

// error handling lol
func errorCheck(err error) {
	if err != nil {
		panic(err)
	}
}
