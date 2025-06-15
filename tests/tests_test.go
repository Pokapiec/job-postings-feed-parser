package tests

import (
	"testing"
	"theladdersexc/ingestion"
	"theladdersexc/models"
)

func TestSalariesStructure(t *testing.T) {
	type TestSalaryIngestInput struct {
		rawInput any
		location models.Location
		output   models.Salary
	}

	testData := []TestSalaryIngestInput{
		{rawInput: 80_000, location: models.Location{Country: "UK"}, output: models.Salary{Value: float64(80_000), Currency: "GBP"}},
	}

	ingester := ingestion.NewJobIngester()
	for _, data := range testData {
		output, err := ingester.NormalizeSalary(data.rawInput, data.location)
		if err != nil {
			t.Errorf("Unexpected error when normalizing salary: %s", err)
			return
		}

		if output != data.output {
			t.Errorf("Value %v is not equal %v", output, data.output)
			return
		}
	}
}
