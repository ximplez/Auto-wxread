package io

import (
	"io"
	"os"
)

func WriteFile(filePath string, data string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	_ = file.Chmod(0755)
	defer file.Close() // 关闭文件

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}

	return nil
}

func WriteFileByByte(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close() // 关闭文件

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
