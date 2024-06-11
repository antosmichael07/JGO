package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	lgr "github.com/antosmichael07/Go-Logger"
)

var logger = lgr.NewLogger("JGO")

func main() {
	logger.Output.File = false

	cmd := exec.Command("dir", "*.jgo")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Log(lgr.Error, "Error running command \"%s\"", err)
		return
	}
	files := strings.Split(out.String(), " ")
	files[len(files)-1] = files[len(files)-1][:len(files[len(files)-1])-1]

	for _, v := range files {
		logger.Log(lgr.Info, "Start compiling \"%s\"", v)

		data, err := os.ReadFile(v)
		if err != nil {
			logger.Log(lgr.Error, "Error reading file \"%s\", error \"%s\"", v, err)
			return
		}

		lines := strings.Split(string(data), "\n")

		for i := range lines {
			if strings.Index(lines[i], "func") == 0 {
				lines[i] = strings.Replace(lines[i], "func", "function", 1)
			}
		}

		for i, v := range lines {
			fmt.Printf("%d | %s\n", i + 1, v)
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
			return
		}
	}
}
