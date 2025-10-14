package io

import (
	"bufio"
	"io"
	"os"

	"github.com/ximplez/wxread/json_tool"
)

func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // 关闭文件

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ReadJsonFile[T any](filePath string) (*T, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // 关闭文件

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return json_tool.PhaseJson[T](data), nil
}

func ScanFile(filePath string, handler func(string) error) error {
	// open file
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	// remember to close the file at the end of the program
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		// do something with a word
		if err := handler(scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
