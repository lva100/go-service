package export

import (
	"fmt"
	"strconv"

	"github.com/lva100/go-service/internal/models"
	"github.com/xuri/excelize/v2"
)

func GenerateXLS(data []models.Otkrep) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	/*
		index, err := f.NewSheet("Sheet2")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	*/
	// Create a new style for the row
	styleID, err := f.NewStyle(&excelize.Style{
		// Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFFFCC"}}, // Light yellow background
		Font: &excelize.Font{Bold: true, Color: "000000"},
	})
	if err != nil {
		fmt.Println(err)
	}
	// Set the style for row 1
	if err := f.SetRowStyle("Sheet1", 1, 1, styleID); err != nil {
		fmt.Println(err)
	}

	f.SetCellValue("Sheet1", "A1", "ЕНП")
	f.SetCellValue("Sheet1", "B1", "Код МО новый")
	f.SetCellValue("Sheet1", "C1", "Наименование МО новое")
	f.SetCellValue("Sheet1", "D1", "Дата прикрепления")
	f.SetCellValue("Sheet1", "E1", "Дата открепления")
	f.SetCellValue("Sheet1", "F1", "Код МО")

	for index, value := range data {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa((index+2)), value.ENP)
		f.SetCellValue("Sheet1", "B"+strconv.Itoa((index+2)), value.LpuCodeNew)
		f.SetCellValue("Sheet1", "C"+strconv.Itoa((index+2)), value.LpuNameNew)
		f.SetCellValue("Sheet1", "D"+strconv.Itoa((index+2)), value.LpuStart.Format("02.01.2006"))
		f.SetCellValue("Sheet1", "E"+strconv.Itoa((index+2)), value.LpuFinish.Format("02.01.2006"))
		f.SetCellValue("Sheet1", "F"+strconv.Itoa((index+2)), value.LpuCode)
	}

	f.SetColWidth("Sheet1", "A", "A", 20)
	f.SetColWidth("Sheet1", "B", "B", 15)
	f.SetColWidth("Sheet1", "C", "C", 60)
	f.SetColWidth("Sheet1", "D", "D", 20)
	f.SetColWidth("Sheet1", "E", "E", 20)
	f.SetColWidth("Sheet1", "F", "F", 10)

	return f, nil
}
