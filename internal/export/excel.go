package export

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/lva100/go-service/internal/models"
	"github.com/xuri/excelize/v2"
)

type Excel struct {
}

func NewExcel() *Excel {
	return &Excel{}
}

func (e *Excel) Generate(data []models.UslReport) (*bytes.Buffer, error) {
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

	f.SetCellValue("Sheet1", "A1", "Период")
	f.SetCellValue("Sheet1", "B1", "Код МО")
	f.SetCellValue("Sheet1", "C1", "Наименование МО")
	f.SetCellValue("Sheet1", "D1", "Код услуги")
	f.SetCellValue("Sheet1", "E1", "МС")
	f.SetCellValue("Sheet1", "F1", "MF")
	f.SetCellValue("Sheet1", "G1", "Объемы услуг")
	f.SetCellValue("Sheet1", "H1", "Стоимость услуг")

	for index, value := range data {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa((index+2)), value.Start.Format("2006-01-02"))
		// f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), val)
		f.SetCellValue("Sheet1", "B"+strconv.Itoa((index+2)), value.Code_MO)
		f.SetCellValue("Sheet1", "C"+strconv.Itoa((index+2)), value.OrgName)
		f.SetCellValue("Sheet1", "D"+strconv.Itoa((index+2)), value.Code_Usl)
		f.SetCellValue("Sheet1", "E"+strconv.Itoa((index+2)), value.MC)
		f.SetCellValue("Sheet1", "F"+strconv.Itoa((index+2)), value.MF)
		f.SetCellValue("Sheet1", "G"+strconv.Itoa((index+2)), value.Usl_vol)
		f.SetCellValue("Sheet1", "H"+strconv.Itoa((index+2)), value.Usl_fin)
	}

	// fmt.Println(len(data))

	if err = f.SetCellFormula("Sheet1", fmt.Sprintf("G%d", len(data)+2), fmt.Sprintf("=SUM(G2:G%d)", len(data)+1)); err != nil {
		fmt.Println(err)
	}
	if err = f.SetCellFormula("Sheet1", fmt.Sprintf("H%d", len(data)+2), fmt.Sprintf("=SUM(H2:H%d)", len(data)+1)); err != nil {
		fmt.Println(err)
	}

	if err := f.SetRowStyle("Sheet1", len(data)+2, len(data)+2, styleID); err != nil {
		fmt.Println(err)
	}

	// Calculate and set column width (manual estimation)
	/*
		maxWidthColA := len("Период")
		maxWidthColB := len("Код МО")
		maxWidthColC := len("Наименование МО")
		maxWidthColD := len("Код услуги")
		maxWidthColE := len("МС")
		maxWidthColF := len("MF")
		maxWidthColG := len("Объемы услуг")
		maxWidthColH := len("Стоимость услуг")
		for _, val := range data {
			if len(val.Start.Format("2006-01-02")) > maxWidthColA {
				maxWidthColA = len(val.Start.Format("2006-01-02"))
			}
			if len(val.Code_MO) > maxWidthColB {
				maxWidthColB = len(val.Code_MO)
			}
			if len(val.OrgName) > maxWidthColC {
				maxWidthColC = len(val.OrgName)
			}
			if len(val.Code_Usl) > maxWidthColD {
				maxWidthColD = len(val.Code_Usl)
			}
			if len(fmt.Sprintf("%d", val.MC)) > maxWidthColE {
				maxWidthColE = len(fmt.Sprintf("%d", val.MC))
			}
			if len(fmt.Sprintf("%d", val.MF)) > maxWidthColF {
				maxWidthColF = len(fmt.Sprintf("%d", val.MF))
			}
			if len(fmt.Sprintf("%f", val.Usl_vol)) > maxWidthColG {
				maxWidthColG = len(fmt.Sprintf("%f", val.Usl_vol))
			}
			if len(fmt.Sprintf("%f", val.Usl_fin)) > maxWidthColH {
				maxWidthColH = len(fmt.Sprintf("%f", val.Usl_fin))
			}
		}

		// Estimate character width (adjust based on your font/size)
		// This is a rough estimate; you might need to fine-tune it.
		estimatedCharWidth := 1.5 // Adjust as needed
		calculatedColAWidth := float64(maxWidthColA) * estimatedCharWidth
		calculatedColBWidth := float64(maxWidthColB) * estimatedCharWidth
		calculatedColCWidth := float64(maxWidthColC) * estimatedCharWidth
		calculatedColDWidth := float64(maxWidthColD) * estimatedCharWidth
		calculatedColEWidth := float64(maxWidthColE) * estimatedCharWidth
		calculatedColFWidth := float64(maxWidthColF) * estimatedCharWidth
		calculatedColHWidth := float64(maxWidthColH) * estimatedCharWidth
		calculatedColGWidth := float64(maxWidthColG) * estimatedCharWidth
	*/
	// Add some padding
	// calculatedColWidth += 2.0

	f.SetColWidth("Sheet1", "A", "A", 10)
	f.SetColWidth("Sheet1", "B", "B", 8)
	f.SetColWidth("Sheet1", "C", "C", 100)
	f.SetColWidth("Sheet1", "D", "D", 12)
	f.SetColWidth("Sheet1", "E", "E", 4)
	f.SetColWidth("Sheet1", "F", "F", 4)
	f.SetColWidth("Sheet1", "G", "G", 14)
	f.SetColWidth("Sheet1", "H", "H", 16)

	out, _ := f.WriteToBuffer()
	return out, nil
}
