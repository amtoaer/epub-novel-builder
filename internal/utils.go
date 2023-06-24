package internal

import (
	"bytes"
	"io"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(input io.Reader) ([]byte, error) {
	reader := transform.NewReader(input, simplifiedchinese.GBK.NewDecoder())
	return io.ReadAll(reader)
}

func Utf8ToGbk(input io.Reader) ([]byte, error) {
	reader := transform.NewReader(input, simplifiedchinese.GBK.NewEncoder())
	return io.ReadAll(reader)
}

func GbkToUtf8String(input io.Reader) (string, error) {
	content, err := GbkToUtf8(input)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func Utf8ToGbkString(input io.Reader) (string, error) {
	content, err := Utf8ToGbk(input)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GbkStringToUtf8(input string) (string, error) {
	return GbkToUtf8String(bytes.NewBuffer([]byte(input)))
}

func Utf8StringToGbk(input string) (string, error) {
	return Utf8ToGbkString(bytes.NewBuffer([]byte(input)))
}
