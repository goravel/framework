package file

import (
	"bufio"
	"io"
	"os"

	"github.com/goravel/framework/errors"
)

func GetLineNum(file string) int {
	total := 0
	f, _ := os.OpenFile(file, os.O_RDONLY, 0444)
	buf := bufio.NewReader(f)

	for {
		_, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				total++

				break
			}
		} else {
			total++
		}
	}

	defer errors.Ignore(f.Close())

	return total
}
