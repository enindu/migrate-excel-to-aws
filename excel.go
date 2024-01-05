package main

import "github.com/xuri/excelize/v2"

func getFileRows(f *string, s *string) [][]string {
	file, exception := openFile(f)
	handle(exception)

	fileRows, exception := file.GetRows(*s)
	handle(exception)

	return fileRows
}

func openFile(f *string) (*excelize.File, error) {
	return excelize.OpenFile(*f)
}
