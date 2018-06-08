package targets

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// textFileInvalidEntry is a regular expression to match an invalid
// textfile line.
var textFileInvalidEntry = regexp.MustCompile(`[^0-9A-Za-z\-:\._\[\]]+`)

// TextFile represents a textfile target driver.
type TextFile struct {
	File string `mapstructure:"file"`
}

// NewTextFile will return a TextFile.
func NewTextFile(options map[string]interface{}) (*TextFile, error) {
	var textfile TextFile

	err := mapstructure.Decode(options, &textfile)
	if err != nil {
		return nil, err
	}

	if textfile.File == "" {
		return nil, fmt.Errorf("file is a required option for textfile")
	}

	if v, ok := options["_dir"]; ok {
		if dir, ok := v.(string); ok {
			textfile.File = path.Join(dir, textfile.File)
		}
	}

	if _, err := os.Stat(textfile.File); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", textfile.File)
	}

	return &textfile, nil
}

// Discover implements the Target interface for a textfile driver.
// It returns a set of hosts specified in a text file.
func (r TextFile) Discover() ([]Host, error) {
	f, err := os.Open(r.File)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hosts []Host
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "//") {
			continue
		}

		if !textFileInvalidEntry.MatchString(line) {
			host := Host{
				Name:    line,
				Address: line,
			}

			hosts = append(hosts, host)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}
