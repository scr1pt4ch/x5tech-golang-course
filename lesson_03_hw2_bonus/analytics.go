package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

// isValid проверяет корректность записи
func isValid(s Sale) bool {
	return "" != s.Product &&
		"" != s.Category &&
		0 < s.Quantity &&
		0 < s.UnitPriceKopecks
}

// topProduct возвращает товар с наибольшим числом проданных единиц
// если таких несколько - лексикографически наименьшее среди них
func topProduct(units map[string]int) string {
	best := ""
	bestUnits := 0

	for prod, n := range units {
		if bestUnits < n || (n == bestUnits && prod < best) {
			best = prod
			bestUnits = n
		}
	}

	return best
}

func AnalyzeSales(sales []Sale) Report {
	report := Report{
		RevenueByCategory: make(map[string]int64),
	}

	unitsByProduct := make(map[string]int)

	for _, s := range sales {
		if s.Cancelled {
			continue
		}

		if !isValid(s) {
			report.InvalidSales++
			continue
		}

		revenue := int64(s.Quantity) * s.UnitPriceKopecks
		report.RevenueKopecks += revenue
		report.UnitsSold += s.Quantity
		report.RevenueByCategory[s.Category] += revenue
		unitsByProduct[s.Product] += s.Quantity
	}

	report.TopProduct = topProduct(unitsByProduct)
	return report
}
