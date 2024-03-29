package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mgutz/ansi"
)

var ConfigFile string
var InputFile string
var CaseInsensitive bool

type rule struct {
	Name    string `json:"name"`
	Regex   regex  `json:"expr"`
	Color   string `json:"color"`
	Replace string `json:"replace,omitempty"`

	colorFunc func(string) string
}

type regex struct {
	string string
	expr   *regexp.Regexp
}

func (regex *regex) fromString(data string) (err error) {
	regex.string = data
	regex.expr, err = regexp.Compile(data)
	return
}

func (regex *regex) UnmarshalJSON(data []byte) (err error) {
	regex.string = string(data)
	var exprString string
	if err = json.Unmarshal(data, &exprString); err != nil {
		return
	}
	regex.expr, err = regexp.Compile(exprString)
	return
}
func (regex *regex) MarshalJSON() ([]byte, error) {
	return json.Marshal(regex.string)
}

type rules []rule

func (r rule) Transform(line string) string {
	return r.Regex.expr.ReplaceAllStringFunc(line, func(s string) string {
		if r.Replace != "" {
			s = r.Regex.expr.ReplaceAllString(s, r.Replace)
		}
		if r.Color != "" {
			if r.colorFunc == nil {
				r.colorFunc = ansi.ColorFunc(r.Color)
			}
			s = r.colorFunc(s)
		}
		return s
	})
}

func readRules(cfgFileName string) ([]rule, error) {
	if file, err := os.Open(cfgFileName); err != nil {
		return nil, err
	} else if data, err := ioutil.ReadAll(file); err != nil {
		return nil, err
	} else {
		rules := rules{}
		if err := json.Unmarshal(data, &rules); err != nil {
			return nil, err
		}
		return rules, nil

	}
}

func init() {
	config_file := path.Join("/", "usr", "share", "log-colorizer", "config", "config.json")
	current_user, err := user.Current()
	if err == nil {
		homeDirConfigfile := path.Join(current_user.HomeDir, ".config.json")
		if exists(homeDirConfigfile) {
			config_file = path.Join(current_user.HomeDir, ".config.json")
		}
	}

	flag.StringVar(&ConfigFile, "c", config_file, "config file location")
	flag.StringVar(&InputFile, "f", "", "input file (defaults to stdin)")
	flag.BoolVar(&CaseInsensitive, "i", false, "case insensitive")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getInputFile() (io.ReadCloser, error) {
	if InputFile == "" {
		return ioutil.NopCloser(os.Stdin), nil
	} else if file, err := os.Open(InputFile); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func main() {
	flag.Parse()
	fmt.Printf("Log-colorizer version: %s build time: %s, config: %s\n", Version, BuildTime, ConfigFile)
	rawReader, err := getInputFile()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rawReader.Close()
	if rules, err := readRules(ConfigFile); err != nil {
		fmt.Println("Error while reading rules:", err)
		fmt.Println("Aborting.")
	} else {
		args := flag.Args()
		if len(args) > 0 {
			fmt.Printf("Highlite: %+v\n", args)
			colors := []string{
				"red",
				"green",
				"yellow",
				"blue",
				"magenta",
				"cyan",
			}
			colorLen := len(colors)
			ci := 0
			for i, part := range args {
				if CaseInsensitive {
					part = "(?i)" + part
				}
				partEx, err := regexp.Compile(part)
				if err != nil {
					fmt.Println("Cant use rule:", part, err)
				} else {
					rules = append(rules, rule{
						Name: strconv.Itoa(i),
						Regex: regex{
							string: part,
							expr:   partEx,
						},
						Color: fmt.Sprintf("+b:%s", colors[ci]),
					})
					ci += 1
					if colorLen <= ci {
						ci = 0
					}
				}
			}
		}
		rdr := bufio.NewReaderSize(rawReader, 100)
		read := func() (string, error) {
			return rdr.ReadString('\n')
		}
		for line, err := read(); err != io.EOF; line, err = read() {
			for _, rule := range rules {
				line = rule.Transform(strings.TrimRightFunc(line, unicode.IsSpace))
			}
			fmt.Println(line)

		}
	}

}
