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
	"strings"
	"unicode"

	"github.com/mgutz/ansi"
	"strconv"
)

var CONFIG_FILE string
var INPUT_FILE string

type rule struct {
	Name    string `json:"name"`
	Regex   regex  `json:"expr"`
	Color   string `json:"color"`
	Replace string `json:"replace,omitempty"`
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
			s = ansi.Color(s, r.Color)
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
	config_dir := "./"
	current_user, err := user.Current()
	if err == nil {
		config_dir = current_user.HomeDir
	}

	config_file := path.Join(config_dir, ".config.json")

	flag.StringVar(&CONFIG_FILE, "c", config_file, "config file location")
	flag.StringVar(&INPUT_FILE, "i", "", "input file (defaults to stdin)")
}

func getInputFile() (io.ReadCloser, error) {
	if INPUT_FILE == "" {
		return ioutil.NopCloser(os.Stdin), nil
	} else if file, err := os.Open(INPUT_FILE); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func main() {
	flag.Parse()
	rawReader, err := getInputFile()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rawReader.Close()

	if rules, err := readRules(CONFIG_FILE); err != nil {
		fmt.Println("Error while reading rules:", err)
		fmt.Println("Aborting.")
	} else {
		args := flag.Args()
		fmt.Printf("Got parts: %+v\n", args)
		if len(args) > 0 {
			for i, part := range args {
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
						Color: "red",
					})
				}
			}
		}

		rdr := bufio.NewReader(rawReader)

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
