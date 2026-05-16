package main

import (
	"fmt"
	"strings"
)

// NewLine Separator
func WordsByNewLine(words []string) string {
	var result strings.Builder
	for _, word := range words {
		result.WriteString(word)
		result.WriteString("\n")
	}
	return result.String()

}

// Comma Separator
func WordsByComma(words []string) string {
	var b strings.Builder
	for i, res := range words {
		b.WriteString(res)
		if i < len(words)-1 {
			b.WriteString(",")
		}

	}
	return b.String()
}

// Building CSV row
func buildCSVRow(fields []string) string {
	var result strings.Builder
	for i, field := range fields {
		if strings.Contains(field, ",") {
			result.WriteString(`"`)
			result.WriteString(field)
			result.WriteString(`"`)
		} else {
			result.WriteString(field)
		}

		if i < len(fields) {
			result.WriteString(",")
		}
	}

	return result.String()
}

// Building HTML Table
func buildTable(headers []string, rows [][]string) string {
	var result strings.Builder
	result.WriteString("<table>\n<tr>")

	for _, h := range headers {
		result.WriteString("<th>")
		result.WriteString(h)
		result.WriteString("</th>")
	}

	result.WriteString("</tr>\n")
	for _, row := range rows {
		result.WriteString("<tr>")
		for _, cell := range row {
			result.WriteString("<td>")
			result.WriteString(cell)
			result.WriteString("</td>")

		}
		result.WriteString("</tr>\n")

	}
	result.WriteString("</table>")
	return result.String()

}

// Building SQL query
func buildQuery(table string, conditions map[string]string) string {
	var b strings.Builder

	b.WriteString("SELECT * FROM ")
	b.WriteString(table)

	if len(conditions) > 0 {
		b.WriteString(" WHERE ")
		i := 0
		for col, val := range conditions {
			b.WriteString(col)
			b.WriteString(" = '")
			b.WriteString(val)
			b.WriteString("'")
			if i < len(conditions)-1 {
				b.WriteString(" AND ")
			}
			i++
		}
	}

	return b.String()
}

func formatLog(level string, message string, fields map[string]string) string {
	var result strings.Builder

	result.WriteString("[")
	result.WriteString(level)
	result.WriteString("]")
	result.WriteString(message)

	for k, v := range fields {
		result.WriteString(" ")
		result.WriteString(k)
		result.WriteString("=")
		result.WriteString(v)
	}

	return result.String()
}

func main() {

	numbers := []string{"one", "two", "three", "four", "five", "six"}
	fmt.Println(WordsByNewLine(numbers))

	words := []string{"apple", "banana", "cherry"}
	fmt.Println(WordsByComma(words))

	row := buildCSVRow([]string{"Michael", "Lagos", "25", "Language", "GoLang", "Cloud ,DevOps ,Engineer"})
	fmt.Println(row)

	query := buildQuery("users", map[string]string{
		"city": "Lagos",
		"role": "admin",
	})
	fmt.Println(query)

	log := formatLog("ERROR", "database connection failed", map[string]string{
		"host": "localhost",
		"port": "5432",
	})
	fmt.Println(log)

	headers := []string{"Name", "City", "Age"}
	rows := [][]string{
		{"Michael", "Lagos", "25"},
		{"Alice", "Abuja", "30"},
		{"Bob", "Kano", "22"},
	}
	fmt.Println(buildTable(headers, rows))
	fmt.Println("---")

}
