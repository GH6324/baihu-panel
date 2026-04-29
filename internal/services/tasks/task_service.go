package tasks

import (
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/models"
	"github.com/engigu/baihu-panel/internal/utils"
	"strings"
)

type TaskService struct {
}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (ts *TaskService) GetTaskBySourceID(sourceID string) *models.Task {
	var task models.Task
	res := database.DB.Where("source_id = ?", sourceID).Limit(1).Find(&task)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil
	}
	return &task
}

func (ts *TaskService) CreateTask(name, command, schedule string, timeout int, workDir, cleanConfig, envs, taskType, config string, agentID *string, languages models.TaskLanguages, triggerType string, tags string, retryCount int, retryInterval int, randomRange int, sourceID string, pinType string) *models.Task {
	if taskType == "" {
		taskType = "task"
	}
	if triggerType == "" {
		triggerType = constant.TriggerTypeCron
	}
	if pinType == "" {
		pinType = constant.PinTypeNone
	}
	task := &models.Task{
		ID:            utils.GenerateID(),
		Name:          name,
		Command:       models.BigText(command),
		PinType:       pinType,
		Tags:          tags,
		Type:          taskType,
		TriggerType:   triggerType,
		Config:        models.BigText(config),
		Schedule:      schedule,
		Timeout:       timeout,
		WorkDir:       workDir,
		CleanConfig:   cleanConfig,
		Envs:          models.BigText(envs),
		Languages:     languages,
		AgentID:       agentID,
		Enabled:       utils.BoolPtr(true),
		RetryCount:    retryCount,
		RetryInterval: retryInterval,
		RandomRange:   randomRange,
		SourceID:      sourceID,
		CreatedAt:     models.Now(),
		UpdatedAt:     models.Now(),
	}
	if triggerType != constant.TriggerTypeCron {
		task.NextRun = nil
	}
	database.DB.Select("*").Create(task)


	return task
}

func (ts *TaskService) GetTasks() []models.Task {
	var tasks []models.Task
	database.DB.Find(&tasks)
	return tasks
}

// GetTasksWithPagination 分页获取任务列表
func (ts *TaskService) GetTasksWithPagination(page, pageSize int, name string, agentID *string, tags string, taskType string) ([]models.Task, int64) {
	var tasks []models.Task
	var total int64

	query := database.DB.Model(&models.Task{})
	if name != "" {
		query = query.Where("name LIKE ? OR remark LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	
	// 标签筛选 (并集)
	if tags != "" {
		tagList := strings.Split(tags, ",")
		var orConditions []string
		var orValues []interface{}
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				orConditions = append(orConditions, "tags LIKE ?")
				orValues = append(orValues, "%"+tag+"%")
			}
		}
		if len(orConditions) > 0 {
			query = query.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}

	if taskType != "" && taskType != "all" {
		query = query.Where("type = ?", taskType)
	}
	if agentID != nil {
		query = query.Where("agent_id = ?", *agentID)
	}

	query.Count(&total)
	query.Order("pin_type DESC, created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks)

	return tasks, total
}

func (ts *TaskService) GetTaskByID(id string) *models.Task {
	var task models.Task
	res := database.DB.Where("id = ?", id).Limit(1).Find(&task)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil
	}
	return &task
}

func (ts *TaskService) UpdateTask(id string, name, command, schedule string, timeout int, workDir, cleanConfig, envs string, enabled bool, taskType, config string, agentID *string, languages models.TaskLanguages, triggerType string, tags string, retryCount int, retryInterval int, randomRange int, sourceID string, pinType string) *models.Task {
	var task models.Task
	res := database.DB.Where("id = ?", id).Limit(1).Find(&task)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil
	}
	task.Name = name
	task.Command = models.BigText(command)
	task.PinType = pinType
	task.Tags = tags
	task.Schedule = schedule
	task.Timeout = timeout
	task.WorkDir = workDir
	task.CleanConfig = cleanConfig
	task.Envs = models.BigText(envs)
	task.Enabled = &enabled
	task.AgentID = agentID
	task.Languages = languages
	task.Config = models.BigText(config)
	task.RetryCount = retryCount
	task.RetryInterval = retryInterval
	task.RandomRange = randomRange
	if taskType != "" {
		task.Type = taskType
	}
	if triggerType != "" {
		task.TriggerType = triggerType
	}
	if sourceID != "" {
		task.SourceID = sourceID
	}

	database.DB.Model(&task).Select(
		"Name", "Command", "Tags", "Schedule", "Timeout", "WorkDir",
		"CleanConfig", "Envs", "Enabled", "AgentID", "Languages",
		"RetryCount", "RetryInterval", "RandomRange", "Type",
		"TriggerType", "Config", "SourceID", "PinType",
	).Updates(&task)

	return &task
}

func (ts *TaskService) DeleteTask(id string) bool {
	// 同时删除关联的通知推送设置
	database.DB.Where("type = ? AND data_id = ?", constant.BindingTypeTask, id).Delete(&models.NotifyBinding{})
	
	result := database.DB.Where("id = ?", id).Delete(&models.Task{})
	return result.RowsAffected > 0
}

func (ts *TaskService) BatchDeleteTasks(ids []string) int64 {
	// 同时删除关联的通知推送设置
	database.DB.Where("type = ? AND data_id IN ?", constant.BindingTypeTask, ids).Delete(&models.NotifyBinding{})
	
	result := database.DB.Where("id IN ?", ids).Delete(&models.Task{})
	return result.RowsAffected
}

// GetAllTags 获取所有任务标签
func (ts *TaskService) GetAllTags() ([]string, error) {
	var tasks []models.Task
	database.DB.Select("tags").Where("tags != ?", "").Find(&tasks)

	tagMap := make(map[string]bool)
	for _, task := range tasks {
		tags := strings.Split(task.Tags, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMap[tag] = true
			}
		}
	}

	result := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		result = append(result, tag)
	}
	return result, nil
}
