package storage

import (
	"log"
	"theladdersexc/models"
)

type JobStorage struct {
	approvedJobs []*models.JobPosting
	rejectedJobs []*models.ApprovalResult
}

func NewJobStorage() *JobStorage {
	return &JobStorage{
		approvedJobs: make([]*models.JobPosting, 0),
		rejectedJobs: make([]*models.ApprovalResult, 0),
	}
}

func (js *JobStorage) StoreApprovedJob(job *models.JobPosting) {
	js.approvedJobs = append(js.approvedJobs, job)
	log.Printf("Stored approved job: %s - %s", job.ID, job.Title)
}

func (js *JobStorage) LogRejectedJob(result *models.ApprovalResult) {
	js.rejectedJobs = append(js.rejectedJobs, result)
	log.Printf("Logged rejected job: %s - Reasons: %v", result.JobID, result.Reasons)
}

func (js *JobStorage) GetApprovedJobs() []*models.JobPosting {
	return js.approvedJobs
}

func (js *JobStorage) GetRejectedJobs() []*models.ApprovalResult {
	return js.rejectedJobs
}

func (js *JobStorage) GetStats() (int, int) {
	return len(js.approvedJobs), len(js.rejectedJobs)
}
