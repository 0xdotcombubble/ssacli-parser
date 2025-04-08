package main 
import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings" ) 
	
var (
	usageRG = regexp.MustCompile("^.*: (.*)%")
	dayRG = regexp.MustCompile("^.*date: (.*) ")
	bayRG = regexp.MustCompile("^.*Bay: (.*)")
	boxRG = regexp.MustCompile("^.*Box: (.*)")
	statusRG = regexp.MustCompile("  Status: (.*)")
	ctempRG = regexp.MustCompile("^.*Current Temperature.*: (.*)")
	mtempRG = regexp.MustCompile("^.*Maximum Temperature.*: (.*)")
	powerRG = regexp.MustCompile("^.*Power On Hours: (.*)")
	typeRG = regexp.MustCompile("^.*Interface Type: (.*)") ) 

// parse reads the input content and prints out the metrics in a formatted way. 

func parse(content string) {
	// Default values that will update when matching lines are found.
	bayCurrent := "1"
	boxCurrent := "1"
	diskTypeCurrent := "none"
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		// Update bay, box, and disk type if found.
		if bayMatch := bayRG.FindStringSubmatch(line); len(bayMatch) != 0 {
			bayCurrent = bayMatch[1]
		}
		if boxMatch := boxRG.FindStringSubmatch(line); len(boxMatch) != 0 {
			boxCurrent = boxMatch[1]
		}
		if typeMatch := typeRG.FindStringSubmatch(line); len(typeMatch) != 0 {
			diskTypeCurrent = typeMatch[1]
		}
		// Build a name identifier using the current bay, box, and type values.
		name := "box " + boxCurrent + " bay " + bayCurrent + " type " + diskTypeCurrent
		// Process usage
		if usageMatch := usageRG.FindStringSubmatch(line); len(usageMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(usageMatch[1]), 64); err == nil {
				fmt.Printf("%s: disk_usage_remaining: %f\n", name, metricValue)
			}
		}
		// Process estimated life remaining
		if dayMatch := dayRG.FindStringSubmatch(line); len(dayMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(dayMatch[1]), 64); err == nil {
				fmt.Printf("%s: disk_estimated_life_remaining: %f\n", name, metricValue)
			}
		}
		// Process disk status
		if statusMatch := statusRG.FindStringSubmatch(line); len(statusMatch) != 0 {
			metricValue := 0.0
			if strings.TrimSpace(statusMatch[1]) == "OK" {
				metricValue = 1.0
			}
			fmt.Printf("%s: disk_status: %f\n", name, metricValue)
		}
		// Process current temperature
		if ctempMatch := ctempRG.FindStringSubmatch(line); len(ctempMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(ctempMatch[1]), 64); err == nil {
				fmt.Printf("%s: disk_current_temperature: %f\n", name, metricValue)
			}
		}
		// Process maximum temperature
		if mtempMatch := mtempRG.FindStringSubmatch(line); len(mtempMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(mtempMatch[1]), 64); err == nil {
				fmt.Printf("%s: disk_maximum_temperature: %f\n", name, metricValue)
			}
		}
		// Process power on hours
		if powerMatch := powerRG.FindStringSubmatch(line); len(powerMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(powerMatch[1]), 64); err == nil {
				fmt.Printf("%s: disk_power_on_hours: %f\n", name, metricValue)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading input: %v", err)
	} 
} 

func main() {
	// Define a flag for the input file.
	filePath := flag.String("file", "", "Path to the input file containing disk details")
	flag.Parse()
	if *filePath == "" {
		log.Fatal("Please provide an input file path using the -file flag")
	}
	// Read the entire file content.
	data, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	// Parse and print the metrics from the file content.
	parse(string(data))
}
