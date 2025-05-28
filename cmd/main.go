package main

import (
	"fmt"
	"log"
	"path/filepath"
	"theladdersexc/approval"
	"theladdersexc/ingestion"
	"theladdersexc/storage"
)

type JobProcessingPipeline struct {
	ingester *ingestion.JobIngester
	approver *approval.JobApprover
	storage  *storage.JobStorage
}

func NewJobProcessingPipeline() *JobProcessingPipeline {
	return &JobProcessingPipeline{
		ingester: ingestion.NewJobIngester(),
		approver: approval.NewJobApprover(),
		storage:  storage.NewJobStorage(),
	}
}

func (jpp *JobProcessingPipeline) ProcessJobsFromFile(filePath string) error {
	jobs, err := jpp.ingester.IngestFromFile(filePath)
	if err != nil {
		return fmt.Errorf("ingestion failed for file %s: %w", filePath, err)
	}

	log.Printf("Ingested %d jobs from %s", len(jobs), filepath.Base(filePath))

	for _, job := range jobs {
		result := jpp.approver.ApproveJob(job)

		if result.Approved {
			jpp.storage.StoreApprovedJob(job)
		} else {
			jpp.storage.LogRejectedJob(result)
		}
	}

	return nil
}

func (jpp *JobProcessingPipeline) ProcessJobsFromFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		if err := jpp.ProcessJobsFromFile(filePath); err != nil {
			log.Printf("Error processing file %s: %v", filePath, err)
			continue
		}
	}
	return nil
}

func (jpp *JobProcessingPipeline) ProcessJobs(jsonData []byte) error {
	jobs, err := jpp.ingester.IngestFromJSON(jsonData)
	if err != nil {
		return fmt.Errorf("ingestion failed: %w", err)
	}

	log.Printf("Ingested %d jobs", len(jobs))

	for _, job := range jobs {
		result := jpp.approver.ApproveJob(job)

		if result.Approved {
			jpp.storage.StoreApprovedJob(job)
		} else {
			jpp.storage.LogRejectedJob(result)
		}
	}

	return nil
}

func main() {
	fmt.Println("=== Job Ingestion and Approval System ===")
	fmt.Println()

	feedFiles := []string{"./feed_files/feed1.json", "./feed_files/feed2.json"}

	for _, file := range feedFiles {
		fmt.Printf("  - %s\n", file)
	}
	fmt.Println()

	pipeline := NewJobProcessingPipeline()

	for i, feedFile := range feedFiles {
		fmt.Printf("=== Processing Feed %d: %s ===\n", i+1, filepath.Base(feedFile))
		if err := pipeline.ProcessJobsFromFile(feedFile); err != nil {
			log.Printf("Error processing feed file %s: %v", feedFile, err)
			continue
		}
		fmt.Println()
	}

	storage := pipeline.storage
	approvedCount, rejectedCount := storage.GetStats()

	fmt.Printf("=== FINAL RESULTS ===\n")
	fmt.Printf("Total Approved: %d\n", approvedCount)
	fmt.Printf("Total Rejected: %d\n", rejectedCount)
	fmt.Printf("Total Processed: %d\n", approvedCount+rejectedCount)

	if approvedCount > 0 {
		fmt.Println("\n=== APPROVED JOBS ===")
		// Create a temporary approver instance for the helper function
		tempApprover := approval.NewJobApprover()
		for i, job := range storage.GetApprovedJobs() {
			salaryStr := fmt.Sprintf("%.0f %s", tempApprover.ExtractSalaryValue(job.Salary.Value), job.Salary.Currency)
			if job.Salary.Unit == "hourly" {
				salaryStr += "/hour"
			}

			remoteSuffix := ""
			if job.Remote {
				remoteSuffix = " (Remote)"
			}

			fmt.Printf("%d. ✓ %s\n", i+1, job.Title)
			fmt.Printf("   Company: %s\n", job.Company)
			fmt.Printf("   Location: %s, %s, %s%s\n", job.Location.City, job.Location.State, job.Location.Country, remoteSuffix)
			fmt.Printf("   Salary: %s\n", salaryStr)
			fmt.Printf("   Type: %s\n\n", job.EmploymentType)
		}
	}

	if rejectedCount > 0 {
		fmt.Println("=== REJECTED JOBS ===")
		for i, rejection := range storage.GetRejectedJobs() {
			fmt.Printf("%d. ✗ Job ID: %s\n", i+1, rejection.JobID)
			fmt.Printf("   Rejection Reasons:\n")
			for _, reason := range rejection.Reasons {
				fmt.Printf("     - %s\n", reason)
			}
			fmt.Println()
		}
	}
}
