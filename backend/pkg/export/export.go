package export

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Column struct {
	Header string
	Width  float64
}

func WriteXLSX(w http.ResponseWriter, filename string, cols []Column, rows [][]string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(index)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return err
	}

	for i, col := range cols {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col.Header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		if col.Width > 0 {
			colName, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheet, colName, colName, col.Width)
		}
	}

	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(filename)))

	return f.Write(w)
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer(`"`, `_`, `\`, `_`, `/`, `_`, `:`, `_`, `*`, `_`, `?`, `_`, `<`, `_`, `>`, `_`, `|`, `_`)
	return r.Replace(name)
}
