package main 
import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
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
	typeRG = regexp.MustCompile("^.*Interface Type: (.*)")
	sizeRG = regexp.MustCompile("^\\s*Size: (.*)") 
)

type DiskKey struct {
	Box int
	Bay int
}

// parse reads the input content and prints out the metrics in a formatted way. 
func parse(content string) {
	// Default values that will update when matching lines are found.
	bayCurrent := 1
	boxCurrent := 1
	
	// Maps to store metrics for each disk
	diskMetrics := make(map[DiskKey]map[string]float64)
	// Map to store disk sizes as tags
	diskSizes := make(map[DiskKey]float64)
	
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		// Update bay and box if found.
		if bayMatch := bayRG.FindStringSubmatch(line); len(bayMatch) != 0 {
			if bay, err := strconv.Atoi(strings.TrimSpace(bayMatch[1])); err == nil {
				bayCurrent = bay
			}
		}
		if boxMatch := boxRG.FindStringSubmatch(line); len(boxMatch) != 0 {
			if box, err := strconv.Atoi(strings.TrimSpace(boxMatch[1])); err == nil {
				boxCurrent = box
			}
		}
		
		// Create a unique disk identifier
		diskKey := DiskKey{Box: boxCurrent, Bay: bayCurrent}
		
		// Initialize metrics map for this disk if not exists
		if _, exists := diskMetrics[diskKey]; !exists {
			diskMetrics[diskKey] = make(map[string]float64)
		}
		
		// Process usage
		if usageMatch := usageRG.FindStringSubmatch(line); len(usageMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(usageMatch[1]), 64); err == nil {
				diskMetrics[diskKey]["usage_remaining"] = metricValue
			}
		}
		// Process estimated life remaining
		if dayMatch := dayRG.FindStringSubmatch(line); len(dayMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(dayMatch[1]), 64); err == nil {
				diskMetrics[diskKey]["estimated_life_remaining"] = metricValue
			}
		}
		// Process disk status
		if statusMatch := statusRG.FindStringSubmatch(line); len(statusMatch) != 0 {
			metricValue := 0.0
			if strings.TrimSpace(statusMatch[1]) == "OK" {
				metricValue = 1.0
			}
			diskMetrics[diskKey]["status"] = metricValue
		}
		// Process current temperature
		if ctempMatch := ctempRG.FindStringSubmatch(line); len(ctempMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(ctempMatch[1]), 64); err == nil {
				diskMetrics[diskKey]["current_temperature"] = metricValue
			}
		}
		// Process maximum temperature
		if mtempMatch := mtempRG.FindStringSubmatch(line); len(mtempMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(mtempMatch[1]), 64); err == nil {
				diskMetrics[diskKey]["maximum_temperature"] = metricValue
			}
		}
		// Process power on hours
		if powerMatch := powerRG.FindStringSubmatch(line); len(powerMatch) != 0 {
			if metricValue, err := strconv.ParseFloat(strings.TrimSpace(powerMatch[1]), 64); err == nil {
				diskMetrics[diskKey]["power_on_hours"] = metricValue
			}
		}
		// Process size
		if sizeMatch := sizeRG.FindStringSubmatch(line); len(sizeMatch) != 0 {
			sizeStr := strings.TrimSpace(sizeMatch[1])
			var metricValue float64
			var err error
			
			if strings.HasSuffix(sizeStr, " TB") {
				sizeStr = strings.TrimSuffix(sizeStr, " TB")
				if metricValue, err = strconv.ParseFloat(sizeStr, 64); err == nil {
					metricValue *= 1024 // Convert TB to GB
				}
			} else if strings.HasSuffix(sizeStr, " GB") {
				sizeStr = strings.TrimSuffix(sizeStr, " GB")
				metricValue, err = strconv.ParseFloat(sizeStr, 64)
			}
			
			if err == nil {
				// Store size in the diskSizes map instead of metrics
				diskSizes[diskKey] = metricValue
			} else {
				log.Printf("Error parsing size: %v", err)
			}
		}
	}
	
	// Create a sorted list of disk keys
	var keys []DiskKey
	for key := range diskMetrics {
		keys = append(keys, key)
	}
	
	// Sort keys by Box and then by Bay
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Box != keys[j].Box {
			return keys[i].Box < keys[j].Box
		}
		return keys[i].Bay < keys[j].Bay
	})
	
	// Output in InfluxDB Line Protocol format
	for _, key := range keys {
		metrics := diskMetrics[key]
		
		// Build tags string with box, bay, and size (if available)
		tagsArray := []string{fmt.Sprintf("box=%d", key.Box), fmt.Sprintf("bay=%d", key.Bay)}
		
		// Add size as a tag if available
		if size, exists := diskSizes[key]; exists {
			// Use integer value for size in tags to avoid float precision issues
			tagsArray = append(tagsArray, fmt.Sprintf("size_gb=%d", int(size)))
		}
		
		// Join all tags
		tags := strings.Join(tagsArray, ",")
		
		// Get sorted metric names
		var metricNames []string
		for metric := range metrics {
			metricNames = append(metricNames, metric)
		}
		sort.Strings(metricNames)
		
		// Build fields in alphabetical order
		var fields []string
		for _, metric := range metricNames {
			fields = append(fields, fmt.Sprintf("%s=%f", metric, metrics[metric]))
		}
		
		// If we have fields, print the line
		if len(fields) > 0 {
			fieldStr := strings.Join(fields, ",")
			fmt.Printf("disk,%s %s\n", tags, fieldStr)
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("error reading input: %v", err)
	} 
} 

func main() {
	// Define a flag for the input file.
	filePath := flag.String("file", "", "Path to the input file containing disk details (leave empty to read from stdin)")
	flag.Parse()
	
	var data []byte
	var err error
	
	if *filePath == "" {
		// No file specified, read from stdin
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("error reading from stdin: %v", err)
		}
	} else {
		// Read from the specified file
		data, err = os.ReadFile(*filePath)
		if err != nil {
			log.Fatalf("error reading file: %v", err)
		}
	}
	
	// Parse and print the metrics from the content.
	parse(string(data))
}
