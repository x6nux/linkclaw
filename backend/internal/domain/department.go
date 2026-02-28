package domain

import "time"

type Department struct {
	ID              string    `gorm:"column:id"                json:"id"`
	CompanyID       string    `gorm:"column:company_id"        json:"companyId"`
	Name            string    `gorm:"column:name"              json:"name"`
	Slug            string    `gorm:"column:slug"              json:"slug"`
	Description     string    `gorm:"column:description"       json:"description"`
	DirectorAgentID *string   `gorm:"column:director_agent_id" json:"directorAgentId"`
	ParentDeptID    *string   `gorm:"column:parent_dept_id"    json:"parentDeptId"`
	CreatedAt       time.Time `gorm:"column:created_at"        json:"createdAt"`
}
