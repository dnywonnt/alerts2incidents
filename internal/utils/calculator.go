package utils // dnywonnt.me/alerts2incidents/internal/utils

import (
	"math"
	"time"
)

// CalculatePages computes the total number of pages needed given the total number of items and the page size.
func CalculatePages(totalItems int, pageSize int) int {
	return int(math.Ceil(float64(totalItems) / float64(pageSize)))
}

// GetCurrentQuarter returns the current quarter of the year.
// It calculates the quarter based on the current month.
func GetCurrentQuarter() int {
	month := int(time.Now().Month()) // Get the current month as an integer
	return (month-1)/3 + 1           // Calculate the quarter from the month
}
