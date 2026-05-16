package main

import (
    "fmt"
    "strings"
)

func buildMarkdownTable(headers []string, rows [][]string) string {
    var result strings.Builder

   
    colWidths := make([]int, len(headers))
    for i, h := range headers {
        colWidths[i] = len(h)
    }


    for _, row := range rows {
        for i, cell := range row {
            if len(cell) > colWidths[i] {
                colWidths[i] = len(cell)
            }
        }
    }

 
    for i, h := range headers {
        result.WriteString("| ")
        result.WriteString(h)
        result.WriteString(strings.Repeat(" ", colWidths[i]-len(h)))
        result.WriteString(" ")
    }
    result.WriteString("|\n")

   
    for _, w := range colWidths {
        result.WriteString("|")
        result.WriteString(strings.Repeat("-", w+2))
    }
    result.WriteString("|\n")

    
    for _, row := range rows {
        for i, cell := range row {
            result.WriteString("| ")
            result.WriteString(cell)
            result.WriteString(strings.Repeat(" ", colWidths[i]-len(cell)))
            result.WriteString(" ")
        }
        result.WriteString("|\n")
    }

    return result.String()
}

func main() {
    
    headers := []string{"Name", "City", "Age"}
    rows := [][]string{
        {"Michael", "Lagos", "25"},
        {"Alice",   "Abuja", "30"},
        {"Bob",     "Kano",  "22"},
    }
    fmt.Println(buildMarkdownTable(headers, rows))
    fmt.Println("---")

    
    headers2 := []string{"Language", "Creator", "Year"}
    rows2 := [][]string{
        {"Go",         "Google",  "2009"},
        {"Python",     "Guido",   "1991"},
        {"JavaScript", "Brendan", "1995"},
    }
    fmt.Println(buildMarkdownTable(headers2, rows2))
    fmt.Println("---")

   
    headers3 := []string{"Fruit"}
    rows3 := [][]string{
        {"Mango"},
        {"Banana"},
        {"Orange"},
    }
    fmt.Println(buildMarkdownTable(headers3, rows3))
}