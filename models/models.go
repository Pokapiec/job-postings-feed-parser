package models

import "time"

type Location struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type Salary struct {
	Value    interface{} `json:"value"` // Can be int, float64, or string
	Currency string      `json:"currency"`
	Unit     string      `json:"unit,omitempty"` // "hourly" or empty for annual
}

// RawJobPosting represents the incoming job data in various formats
type RawJobPosting struct {
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	Company        string      `json:"company"`
	Location       interface{} `json:"location"` // Can be Location struct or string
	Salary         interface{} `json:"salary"`   // Can be Salary struct or number
	EmploymentType string      `json:"employment_type"`
	PostingDate    string      `json:"posting_date"`
	CompanyType    string      `json:"company_type"`
	Language       string      `json:"language"`
	Remote         bool        `json:"remote"`
}

// JobPosting represents the normalized internal job posting structure
type JobPosting struct {
	ID             string
	Title          string
	Description    string
	Company        string
	Location       Location
	Salary         Salary
	EmploymentType string
	PostingDate    time.Time
	CompanyType    string
	Language       string
	Remote         bool
}

// ApprovalResult represents the result of job approval process
type ApprovalResult struct {
	JobID       string
	Approved    bool
	Reasons     []string
	ProcessedAt time.Time
}
