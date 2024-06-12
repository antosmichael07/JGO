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

	if len(os.Args) == 1 {
		fmt.Print("\n!error: no commands called, manual:\n\nThe JGO manual\njgo <command> [options]\n\nCommands:\n\thelp\n\tcheck <directory>\n\tbuild <directory>\n\n")
		return
	}

	if os.Args[1] == "help" && len(os.Args) >= 2 {
		fmt.Print("\nThe JGO manual\njgo <command> [options]\n\nCommands:\n\thelp\n\tcheck <directory>\n\tbuild <directory>\n\n")
		return
	}

	if os.Args[1] == "check" {
		if len(os.Args) >= 3 {
			dir := os.Args[2]
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
				if len(tmp[i]) > 4 && tmp[i][len(tmp[i])-4:] == ".jgo" {
					files = append(files, tmp[i])
				}
			}

			if len(files) != 0 {
				logger.Log(lgr.Info, "JGO files found: %v", files)
			} else {
				logger.Log(lgr.Info, "No JGO files found")
			}
		} else {
			fmt.Print("\n!error: this command requires a directory\n\n")
		}
		return
	}

	if os.Args[1] == "build" {
		if len(os.Args) >= 3 {
			dir := os.Args[2]
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
				if len(tmp[i]) > 4 && tmp[i][len(tmp[i])-4:] == ".jgo" {
					files = append(files, tmp[i])
				}
			}

			if len(files) == 0 {
				logger.Log(lgr.Warning, "No JGO files found")
			}

			for _, v := range files {
				data, err := os.ReadFile(v)
				if err != nil {
					logger.Log(lgr.Error, "Error reading file \"%s\", error \"%s\"", v, err)
					break
				}
				lines := strings.Split(string(data), "\n")

				if strings.Index(lines[0], "package main") != 0 {
					logger.Log(lgr.Info, "Skipping \"%s\", it is not a main package", v)
					continue
				}

				logger.Log(lgr.Info, "Start compiling \"%s\"", v)

				has_main := false

				end := len(lines)
				for i := 0; i < end; i++ {
					// Package
					if strings.Index(lines[i], "package") == 0 {
						lines = append(lines[:i], lines[i+1:]...)
						end--
						for lines[i] == "" || lines[i] == "\r" {
							lines = append(lines[:i], lines[i+1:]...)
							end--
						}
					}

					// Constants
					if strings.Index(lines[i], "const") == 0 {
						lines = append(lines[:i], lines[i+1:]...)
						end--
						for j := i; lines[j][0] != ')'; j++ {
							lines[j] = fmt.Sprintf("const %s", lines[j][get_space_count(lines[j]):])
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
						end_bracket := strings.LastIndex(lines[i], ")")

						for j := start_bracket + 1; j < end_bracket; j++ {
							if lines[i][j] != ' ' && lines[i][j] != ',' && lines[i][j+1] == ' ' && lines[i][j+2] != ' ' && lines[i][j+2] != ',' && j+2 != end_bracket {
								j++
								for lines[i][j] != ',' && j != end_bracket {
									lines[i] = fmt.Sprintf("%s%s", lines[i][:j], lines[i][j+1:])
									end_bracket--
								}
							}
						}

						lines[i] = fmt.Sprintf("%s%s", lines[i][:end_bracket], ") {")
					}

					// Variables
					if strings.Contains(lines[i], ":=") && !is_string(lines[i], strings.Index(lines[i], ":=")) {
						if !strings.Contains(lines[i], " range ") || is_string(lines[i], strings.Index(lines[i], " range ")) {
							for j := strings.Index(lines[i], ":=") - 2; j > 0; j-- {
								if lines[i][j] == ' ' {
									lines[i] = fmt.Sprintf("%s let %s", lines[i][:j], lines[i][j+1:])
									break
								}
							}
							lines[i] = strings.Replace(lines[i], ":=", "=", 1)
						} else {
							lines[i] = strings.Replace(lines[i], "for ", "for let ", 1)
							lines[i] = strings.Replace(lines[i], " := ", " = ", 1)
						}
					}

					// Conditions
					if strings.Contains(lines[i], "if") && !is_string(lines[i], strings.Index(lines[i], "if")) && lines[i][strings.Index(lines[i], "if")+2] == ' ' && (strings.Index(lines[i], "if") == 0 || lines[i][strings.Index(lines[i], "if")-1] == ' ') {
						lines[i] = strings.Replace(lines[i], "if ", "if (", 1)
						lines[i] = strings.Replace(lines[i], " {", ") {", 1)
					}

					// For
					is_for_range := false
					if strings.Contains(lines[i], "for") && !is_string(lines[i], strings.Index(lines[i], "for")) && lines[i][strings.Index(lines[i], "for")+3] == ' ' && (strings.Index(lines[i], "for") == 0 || lines[i][strings.Index(lines[i], "for")-1] == ' ') {
						lines[i] = strings.Replace(lines[i], "for ", "for (", 1)

						if strings.Contains(lines[i], "range ") && !is_string(lines[i], strings.Index(lines[i], "range ")) {
							var_name := ""
							value_name := ""
							var_index := strings.Index(lines[i], "let ")
							for j := var_index + 4; j < len(lines[i]); j++ {
								if lines[i][j] == ' ' || lines[i][j] == ',' {
									var_name = lines[i][var_index+4 : j]
									if lines[i][j] == ',' {
										value_name = lines[i][j+1 : strings.Index(lines[i], " =")]
										lines[i] = fmt.Sprintf("%s%s", lines[i][:j], lines[i][strings.Index(lines[i], " ="):])
										break
									}
									break
								}
							}

							range_index := strings.Index(lines[i], "range ")
							last_str := lines[i][range_index+6:]
							lines[i] = fmt.Sprintf("%s%s%s%s", lines[i][:range_index], "0; ", var_name, " < ")
							for j := 0; j < len(last_str); j++ {
								if last_str[j] == ' ' {
									lines[i] = fmt.Sprintf("%s%s%s%s%s", lines[i], last_str[:j], ".length; ", var_name, "++) {")
									break
								}
							}

							if value_name != "" {
								lines = append(lines[:i+1], append([]string{""}, lines[i+1:]...)...)

								for j := 0; j < get_space_count(lines[i])+3; j++ {
									lines[i+1] = fmt.Sprintf("%s%s", lines[i+1], " ")
								}

								lines[i+1] = fmt.Sprintf("%s%s%s%s%s%s%s", lines[i+1], value_name, " := ", last_str[:strings.Index(last_str, " {")], "[", var_name, "]")
							}

							is_for_range = true
						} else {
							lines[i] = strings.Replace(lines[i], " {", ") {", 1)
						}
					}

					// While
					if !is_for_range && strings.Contains(lines[i], "for") && !is_string(lines[i], strings.Index(lines[i], "for")) && lines[i][strings.Index(lines[i], "for")+3] == ' ' && (strings.Index(lines[i], "for") == 0 || lines[i][strings.Index(lines[i], "for")-1] == ' ') {
						semicolon_count := 0
						for j := 0; j < len(lines[i]); j++ {
							if lines[i][j] == ';' {
								semicolon_count++
							}
						}
						if semicolon_count == 0 {
							lines[i] = strings.Replace(lines[i], "for", "while", 1)
						}
					}

					// Print
					if strings.Contains(lines[i], "fmt.Print") && !is_string(lines[i], strings.Index(lines[i], "fmt.Print")) && lines[i][strings.Index(lines[i], "fmt.Print")+len("fmt.Print")] == '(' && (strings.Index(lines[i], "fmt.Print") == 0 || lines[i][strings.Index(lines[i], "fmt.Print")-1] == ' ') {
						lines[i] = strings.Replace(lines[i], "fmt.Print", "console.log", 1)
					}

					// Array
					if strings.Contains(lines[i], "[]") && !is_string(lines[i], strings.Index(lines[i], "[]")) {
						j := strings.Index(lines[i], "[]")
						for lines[i][j] != '{' {
							lines[i] = fmt.Sprintf("%s%s", lines[i][:j], lines[i][j+1:])
						}
						lines[i] = fmt.Sprintf("%s%s%s", lines[i][:j], "[", lines[i][j+1:])
						for k := j + 1; k < len(lines[i]); k++ {
							if lines[i][k] == '}' {
								lines[i] = fmt.Sprintf("%s%s%s", lines[i][:k], "]", lines[i][k+1:])
								break
							}
						}
					}
				}

				// Call main
				if has_main {
					if lines[len(lines)-1] == "" || lines[len(lines)-1] == "\r" {
						lines = append(lines, "main()")
					} else {
						lines = append(lines, "")
						lines = append(lines, "main()")
					}
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
		} else {
			fmt.Print("\n!error: this command requires a directory\n\n")
		}
		return
	}

	fmt.Print("\n!error: invalid command\n\n")
}

func is_string(str string, index int) bool {
	string_indexes := []int{}
	for i := 0; i < len(str); i++ {
		if str[i] == '"' {
			string_indexes = append(string_indexes, i)
		}
	}
	for i := 0; i < len(string_indexes); i += 2 {
		if index > string_indexes[i] && index < string_indexes[i+1] {
			return true
		}
	}
	return false
}

func get_space_count(str string) int {
	count := 0
	for i := 0; str[i] == ' '; i++ {
		count++
	}
	return count
}
