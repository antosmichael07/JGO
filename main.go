package main

import (
	"fmt"
	"os"
	"strings"

	lgr "github.com/antosmichael07/Go-Logger"
)

var logger = lgr.NewLogger("JGO")

func main() {
	logger.Output.File = false

	dir := os.Args[1]
	out, err := os.ReadDir(dir)
	if err != nil {
		logger.Log(lgr.Error, "Error reading directory \"%s\", error \"%s\"", dir, err)
		return
	}
	tmp := []string{}
	for _, v := range out {
		tmp = append(tmp, v.Name())
	}
	files := []string{}
	for i := range tmp {
		if tmp[i][len(tmp[i])-4:] == ".jgo" {
			files = append(files, tmp[i])
		}
	}

	for _, v := range files {
		logger.Log(lgr.Info, "Start compiling \"%s\"", v)

		has_main := false

		data, err := os.ReadFile(v)
		if err != nil {
			logger.Log(lgr.Error, "Error reading file \"%s\", error \"%s\"", v, err)
			break
		}

		lines := strings.Split(string(data), "\n")

		end := len(lines)
		for i := 0; i < end; i++ {
			// Package
			if strings.Index(lines[i], "package") == 0 {
				lines = append(lines[:i], lines[i+1:]...)
				end--
			}

			// Constants
			if strings.Index(lines[i], "const") == 0 {
				lines = append(lines[:i], lines[i+1:]...)
				end--
				for j := i; lines[j][0] != ')'; j++ {
					char_index := 0
					for lines[j][char_index] == ' ' {
						char_index++
					}
					lines[j] = fmt.Sprintf("const %s", lines[j][char_index:])
					i++
				}
				lines = append(lines[:i], lines[i+1:]...)
				end--
			}

			// Function
			if strings.Index(lines[i], "func") == 0 {
				if strings.Index(lines[i], "main") == 5 {
					has_main = true
				}

				lines[i] = strings.Replace(lines[i], "func", "function", 1)

				start_bracket := strings.Index(lines[i], "(")
				end_bracket := strings.Index(lines[i], ")")

				for j := start_bracket + 1; j < end_bracket; j++ {
					if lines[i][j] != ' ' && lines[i][j] != ',' && lines[i][j] != ')' && lines[i][j+1] == ' ' && lines[i][j+2] != ' ' && lines[i][j+2] != ',' && lines[i][j+2] != ')' && lines[i][j+2] != '{' {
						j++
						for lines[i][j] != ',' && lines[i][j] != ')' {
							lines[i] = fmt.Sprintf("%s%s", lines[i][:j], lines[i][j+1:])
							end_bracket--
						}
					}
				}
			}
		}

		if has_main {
			lines = append(lines, "main()")
		}

		for i, v := range lines {
			space := ""
			for j := 0; j < len(fmt.Sprint(len(lines)))-len(fmt.Sprint(i+1)); j++ {
				space = fmt.Sprintf("%s ", space)
			}
			fmt.Printf("%s%d | %s\n", space, i+1, v)
		}

		to_write := ""
		for _, v := range lines {
			to_write += v
			to_write += "\n"
		}
		to_write = to_write[:len(to_write)-1]

		err = os.WriteFile(fmt.Sprintf("%s%s", v[:len(v)-4], ".js"), []byte(to_write), 0666)
		if err != nil {
			logger.Log(lgr.Error, "Error writing to file \"%s\", error \"%s\"", v, err)
			break
		}
	}
}
