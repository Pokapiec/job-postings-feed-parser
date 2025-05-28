package approval

import (
	"strconv"
	"strings"
	"theladdersexc/models"
	"time"
)

type ApprovalCriteria struct {
	MinAnnualSalaryUSD float64
	MinHourlySalaryUSD float64
}

type JobApprover struct {
	criteria ApprovalCriteria
}

func NewJobApprover() *JobApprover {
	return &JobApprover{
		criteria: ApprovalCriteria{
			MinAnnualSalaryUSD: 100000,
			MinHourlySalaryUSD: 45,
		},
	}
}

func (ja *JobApprover) ApproveJob(job *models.JobPosting) *models.ApprovalResult {
	result := &models.ApprovalResult{
		JobID:       job.ID,
		Approved:    true,
		Reasons:     []string{},
		ProcessedAt: time.Now(),
	}

	// Check geographical location
	if !ja.isValidLocation(job) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "Invalid location: must be remote or in US/Canada")
	}

	// Check employment type
	if !ja.isFullTime(job) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "Invalid employment type: must be full-time")
	}

	// Check salary requirement
	if !ja.meetsSalaryRequirement(job) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "Insufficient salary: must be >$100k annual or >$45/hour")
	}

	// Check company type
	if ja.isStaffingFirm(job) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "Invalid company type: staffing firms not allowed")
	}

	// Check language requirement
	if !ja.hasValidLanguage(job) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "Invalid language: must be English or French (for Canada jobs)")
	}

	return result
}

func (ja *JobApprover) isValidLocation(job *models.JobPosting) bool {
	if job.Remote {
		return true // Remote jobs are allowed from anywhere
	}

	country := strings.ToUpper(strings.TrimSpace(job.Location.Country))
	return country == "USA" || country == "US" || country == "UNITED STATES" ||
		country == "CANADA" || country == "CA"
}

func (ja *JobApprover) isFullTime(job *models.JobPosting) bool {
	empType := strings.ToLower(strings.TrimSpace(job.EmploymentType))
	return empType == "full-time" || empType == "fulltime" || empType == "full time"
}

func (ja *JobApprover) meetsSalaryRequirement(job *models.JobPosting) bool {
	salaryValue := ja.ExtractSalaryValue(job.Salary.Value)
	if salaryValue <= 0 {
		return false
	}

	// Convert to USD (simplified - in production you'd use real exchange rates)
	salaryUSD := ja.convertToUSD(salaryValue, job.Salary.Currency)

	if job.Salary.Unit == "hourly" {
		return salaryUSD >= ja.criteria.MinHourlySalaryUSD
	} else {
		return salaryUSD >= ja.criteria.MinAnnualSalaryUSD
	}
}

func (ja *JobApprover) ExtractSalaryValue(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func (ja *JobApprover) convertToUSD(amount float64, currency string) float64 {
	// Simplified currency conversion - in production, use real exchange rates
	currency = strings.ToUpper(strings.TrimSpace(currency))
	switch currency {
	case "USD":
		return amount
	case "CAD":
		return amount * 0.74 // Approximate CAD to USD
	case "EUR":
		return amount * 1.09 // Approximate EUR to USD
	case "GBP":
		return amount * 1.27 // Approximate GBP to USD
	default:
		return amount // Default to same value if currency unknown
	}
}

func (ja *JobApprover) isStaffingFirm(job *models.JobPosting) bool {
	companyType := strings.ToLower(strings.TrimSpace(job.CompanyType))
	return strings.Contains(companyType, "staffing") || strings.Contains(companyType, "recruiting")
}

func (ja *JobApprover) hasValidLanguage(job *models.JobPosting) bool {
	language := strings.ToLower(strings.TrimSpace(job.Language))

	// Empty language is considered invalid
	if language == "" {
		return false
	}

	// English is always valid
	if language == "english" {
		return true
	}

	// French is valid only for Canadian jobs
	if language == "french" || language == "français" {
		country := strings.ToUpper(strings.TrimSpace(job.Location.Country))
		return country == "CANADA" || country == "CA"
	}

	return false
}
