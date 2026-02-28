package api

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

const (
	taskAttachmentMaxFileSize       = 10 << 20 // 10MB
	taskCreateMultipartMemory int64 = 32 << 20
)

var (
	taskAttachmentAllowedExtensions = map[string]struct{}{
		".jpg":    {},
		".jpeg":   {},
		".png":    {},
		".gif":    {},
		".webp":   {},
		".bmp":    {},
		".svg":    {},
		".pdf":    {},
		".doc":    {},
		".docx":   {},
		".xls":    {},
		".xlsx":   {},
		".ppt":    {},
		".pptx":   {},
		".txt":    {},
		".md":     {},
		".csv":    {},
		".zip":    {},
		".tar.gz": {},
	}
	taskAttachmentAllowedMIMETypes = map[string]struct{}{
		"application/pdf":    {},
		"application/msword": {},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
		"application/vnd.ms-excel": {},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
		"application/vnd.ms-powerpoint":                                             {},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
		"text/plain":                   {},
		"text/markdown":                {},
		"text/csv":                     {},
		"application/zip":              {},
		"application/x-zip-compressed": {},
		"application/x-tar":            {},
		"application/gzip":             {},
		"application/x-gzip":           {},
	}
)

type taskHandler struct {
	taskSvc *service.TaskService
}

func (h *taskHandler) list(c *gin.Context) {
	companyID := currentCompanyID(c)
	q := repository.TaskQuery{
		CompanyID: companyID,
		Status:    domain.TaskStatus(c.Query("status")),
		Priority:  domain.TaskPriority(c.Query("priority")),
	}
	if aid := c.Query("assignee_id"); aid != "" {
		q.AssigneeID = aid
	}
	tasks, total, err := h.taskSvc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks, "total": total})
}

func (h *taskHandler) get(c *gin.Context) {
	task, err := h.taskSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *taskHandler) detail(c *gin.Context) {
	task, err := h.taskSvc.GetTaskDetail(c.Request.Context(), c.Param("id"))
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

type createTaskRequest struct {
	Title       string `json:"title"       form:"title"       binding:"required"`
	Description string `json:"description" form:"description"`
	Priority    string `json:"priority"    form:"priority"`
	AssigneeID  string `json:"assignee_id" form:"assignee_id"`
	ParentID    string `json:"parent_id"   form:"parent_id"`
}

func (h *taskHandler) create(c *gin.Context) {
	var req createTaskRequest
	var files []service.TaskUploadFile

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		var err error
		req, files, err = parseCreateTaskMultipart(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Title = strings.TrimSpace(req.Title)
	}

	agent := currentAgent(c)
	agentID := agent.ID
	in := service.CreateTaskInput{
		CompanyID:   agent.CompanyID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    domain.TaskPriority(req.Priority),
		CreatedBy:   &agentID,
		Attachments: files,
	}
	if req.AssigneeID != "" {
		in.AssigneeID = &req.AssigneeID
	}
	if req.ParentID != "" {
		in.ParentID = &req.ParentID
	}
	task, err := h.taskSvc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func parseCreateTaskMultipart(c *gin.Context) (createTaskRequest, []service.TaskUploadFile, error) {
	if err := c.Request.ParseMultipartForm(taskCreateMultipartMemory); err != nil {
		return createTaskRequest{}, nil, fmt.Errorf("invalid multipart form: %w", err)
	}

	req := createTaskRequest{
		Title:       strings.TrimSpace(c.PostForm("title")),
		Description: strings.TrimSpace(c.PostForm("description")),
		Priority:    strings.TrimSpace(c.PostForm("priority")),
		AssigneeID:  strings.TrimSpace(c.PostForm("assignee_id")),
		ParentID:    strings.TrimSpace(c.PostForm("parent_id")),
	}
	if req.AssigneeID == "" {
		req.AssigneeID = strings.TrimSpace(c.PostForm("assignee"))
	}
	if req.ParentID == "" {
		req.ParentID = strings.TrimSpace(c.PostForm("parent_task"))
	}
	if req.Title == "" {
		return createTaskRequest{}, nil, fmt.Errorf("title is required")
	}

	files, err := parseTaskUploadFiles(c)
	if err != nil {
		return createTaskRequest{}, nil, err
	}
	return req, files, nil
}

func parseTaskUploadFiles(c *gin.Context) ([]service.TaskUploadFile, error) {
	form := c.Request.MultipartForm
	if form == nil || len(form.File) == 0 {
		return nil, nil
	}

	files := make([]service.TaskUploadFile, 0)
	for _, headers := range form.File {
		for _, header := range headers {
			file, err := parseTaskUploadFile(header)
			if err != nil {
				return nil, err
			}
			files = append(files, file)
		}
	}
	return files, nil
}

func parseTaskUploadFile(header *multipart.FileHeader) (service.TaskUploadFile, error) {
	if header == nil {
		return service.TaskUploadFile{}, fmt.Errorf("invalid file")
	}
	if header.Size > taskAttachmentMaxFileSize {
		return service.TaskUploadFile{}, fmt.Errorf("file %q exceeds 10MB limit", header.Filename)
	}

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if !isAllowedTaskAttachment(header.Filename, contentType) {
		return service.TaskUploadFile{}, fmt.Errorf("file %q type is not allowed", header.Filename)
	}

	file, err := header.Open()
	if err != nil {
		return service.TaskUploadFile{}, fmt.Errorf("open file %q: %w", header.Filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, taskAttachmentMaxFileSize+1))
	if err != nil {
		return service.TaskUploadFile{}, fmt.Errorf("read file %q: %w", header.Filename, err)
	}
	if int64(len(data)) > taskAttachmentMaxFileSize {
		return service.TaskUploadFile{}, fmt.Errorf("file %q exceeds 10MB limit", header.Filename)
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(taskAttachmentExtension(header.Filename))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return service.TaskUploadFile{
		OriginalFilename: header.Filename,
		Size:             int64(len(data)),
		MimeType:         contentType,
		Content:          data,
	}, nil
}

func isAllowedTaskAttachment(filename, mimeType string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	if _, ok := taskAttachmentAllowedMIMETypes[mimeType]; ok {
		return true
	}

	ext := taskAttachmentExtension(filename)
	_, ok := taskAttachmentAllowedExtensions[ext]
	return ok
}

func taskAttachmentExtension(filename string) string {
	lower := strings.ToLower(strings.TrimSpace(filename))
	if strings.HasSuffix(lower, ".tar.gz") {
		return ".tar.gz"
	}
	return strings.ToLower(filepath.Ext(lower))
}

func (h *taskHandler) delete(c *gin.Context) {
	companyID := currentCompanyID(c)
	if err := h.taskSvc.Delete(c.Request.Context(), c.Param("id"), companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type addTaskCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *taskHandler) addComment(c *gin.Context) {
	var req addTaskCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent := currentAgent(c)
	comment, err := h.taskSvc.AddComment(c.Request.Context(), c.Param("id"), agent.ID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *taskHandler) deleteComment(c *gin.Context) {
	agent := currentAgent(c)
	if err := h.taskSvc.DeleteComment(c.Request.Context(), c.Param("commentId"), agent.ID, agent.CompanyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type addTaskDependencyRequest struct {
	DependsOnID string `json:"depends_on_id" binding:"required"`
}

func (h *taskHandler) addDependency(c *gin.Context) {
	var req addTaskDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dep, err := h.taskSvc.AddDependency(c.Request.Context(), c.Param("id"), req.DependsOnID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dep)
}

func (h *taskHandler) removeDependency(c *gin.Context) {
	if err := h.taskSvc.RemoveDependency(c.Request.Context(), c.Param("id"), c.Param("depId")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *taskHandler) addWatcher(c *gin.Context) {
	agent := currentAgent(c)
	if err := h.taskSvc.AddWatcher(c.Request.Context(), c.Param("id"), agent.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *taskHandler) removeWatcher(c *gin.Context) {
	agent := currentAgent(c)
	if err := h.taskSvc.RemoveWatcher(c.Request.Context(), c.Param("id"), agent.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type updateTaskTagsRequest struct {
	Tags []string `json:"tags"`
}

func (h *taskHandler) updateTags(c *gin.Context) {
	var req updateTaskTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.taskSvc.UpdateTags(c.Request.Context(), c.Param("id"), domain.StringList(req.Tags)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
