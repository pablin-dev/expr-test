package router

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func LoadBlacklistData(path string) (map[string][]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("failed to close file: %v\n", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty blacklist file")
	}

	header := records[0]
	data := make(map[string][]any)
	for _, col := range header {
		data[col] = []any{}
	}

	for _, record := range records[1:] {
		for i, col := range header {
			if i < len(record) {
				val := record[i]
				var typedVal any = val
				// Attempt conversion
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					typedVal = f
				} else if b, err := strconv.ParseBool(val); err == nil {
					typedVal = b
				}
				data[col] = append(data[col], typedVal)
			}
		}
	}
	return data, nil
}
