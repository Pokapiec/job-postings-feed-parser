package ingestion

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"theladdersexc/models"
	"time"
)

type JobIngester struct{}

func NewJobIngester() *JobIngester {
	return &JobIngester{}
}

func (ji *JobIngester) IngestFromJSON(jsonData []byte) ([]*models.JobPosting, error) {
	var rawJobs []models.RawJobPosting
	if err := json.Unmarshal(jsonData, &rawJobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var jobs []*models.JobPosting
	for i, rawJob := range rawJobs {
		job, err := ji.normalizeJob(rawJob, fmt.Sprintf("job_%d_%d", time.Now().Unix(), i))
		if err != nil {
			log.Printf("Warning: Failed to normalize job %d: %v", i, err)
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (ji *JobIngester) IngestFromFile(filePath string) ([]*models.JobPosting, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	log.Printf("Reading job feed from: %s", filePath)
	return ji.IngestFromJSON(jsonData)
}

func (ji *JobIngester) normalizeJob(raw models.RawJobPosting, id string) (*models.JobPosting, error) {
	job := &models.JobPosting{
		ID:             id,
		Title:          raw.Title,
		Description:    raw.Description,
		Company:        raw.Company,
		EmploymentType: raw.EmploymentType,
		CompanyType:    raw.CompanyType,
		Language:       raw.Language,
		Remote:         raw.Remote,
	}

	if raw.PostingDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", raw.PostingDate); err == nil {
			job.PostingDate = parsedDate
		} else {
			job.PostingDate = time.Now() // Default to current time if parsing fails
		}
	} else {
		job.PostingDate = time.Now()
	}

	location, err := ji.normalizeLocation(raw.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize location: %w", err)
	}
	job.Location = location

	// Normalize salary
	salary, err := ji.normalizeSalary(raw.Salary)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize salary: %w", err)
	}
	job.Salary = salary

	return job, nil
}

func (ji *JobIngester) normalizeLocation(locationData interface{}) (models.Location, error) {
	switch loc := locationData.(type) {
	case map[string]interface{}:
		// Handle structured location object
		location := models.Location{}
		if city, ok := loc["city"].(string); ok {
			location.City = city
		}
		if state, ok := loc["state"].(string); ok {
			location.State = state
		}
		if country, ok := loc["country"].(string); ok {
			location.Country = country
		}
		return location, nil
	case string:
		// Handle string location like "New York, NY, USA"
		return ji.parseLocationString(loc), nil
	default:
		return models.Location{}, fmt.Errorf("unsupported location format: %T", locationData)
	}
}

func (ji *JobIngester) parseLocationString(locationStr string) models.Location {
	parts := strings.Split(locationStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	location := models.Location{}

	switch len(parts) {
	case 3:
		// "City, State, Country"
		location.City = parts[0]
		location.State = parts[1]
		location.Country = parts[2]
	case 2:
		// Could be "City, Country" (like "London, UK")
		location.City = parts[0]
		location.Country = parts[1]
	case 1:
		// Just a city
		location.City = parts[0]
	default:
		// More than expected — fallback logic
		location.City = parts[0]
		location.Country = parts[len(parts)-1]
	}

	return location
}

func (ji *JobIngester) normalizeSalary(salaryData interface{}) (models.Salary, error) {
	switch sal := salaryData.(type) {
	case map[string]interface{}:
		// Handle structured salary object
		salary := models.Salary{}

		if value, ok := sal["value"]; ok {
			salary.Value = value
		}
		if currency, ok := sal["currency"].(string); ok {
			salary.Currency = currency
		}
		if unit, ok := sal["unit"].(string); ok {
			salary.Unit = unit
		}

		return salary, nil
	case float64:
		return models.Salary{
			Value:    sal,
			Currency: "USD", // Default assumption
			Unit:     "",    // Annual by default
		}, nil
	case int:
		// If it is lower than minimal wage for country just assume it is hourly ?
		return models.Salary{
			Value:    float64(sal),
			Currency: "USD",
			Unit:     "",
		}, nil
	default:
		return models.Salary{}, fmt.Errorf("unsupported salary format: %T", salaryData)
	}
}
