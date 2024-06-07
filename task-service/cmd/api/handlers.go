package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DaffaJatmiko/task-service/data"
)

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
// TODO: implement the get task by category method using GetTaskByUserId method in models.go
func (app *Config) GetTask(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		UserID int `json:"user_id"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	task, err := app.Models.Task.GetTasksByUserID(requestPayload.UserID)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// log registration
	err = app.logRequest("get task by user id", fmt.Sprintf("get task by user id: %d", requestPayload.UserID))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Get task by user id %d", requestPayload.UserID),
		Data:    task,
	}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) CreateTask(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		UserID      int    `json:"user_id"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	log.Println(requestPayload)

	task := data.Task{
		Name:        requestPayload.Name,
		Description: requestPayload.Description,
		UserID:      requestPayload.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	log.Println(task)

	_, err = app.Models.Task.Insert(task)
	if err != nil {
		app.errorJSON(w, errors.New("unable to create task"), http.StatusBadRequest)
		return
	}

		// log registration
		err = app.logRequest("create task", fmt.Sprintf("%s added", task.Name))
		if err != nil {
			app.errorJSON(w, err)
			return
		}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Created task %s", task.Name),
		Data:    task,
	}

	app.writeJSON(w, http.StatusCreated, payload)
}

func (app *Config) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := app.Models.Task.GetAll()
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Success",
		Data:    tasks,
	}

	app.writeJSON(w, http.StatusOK, payload)
}


func (app *Config) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			UserID      int    `json:"user_id"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
	}

	task := data.Task{
			ID:          requestPayload.ID,
			Name:        requestPayload.Name,
			Description: requestPayload.Description,
			UserID:      requestPayload.UserID,
			UpdatedAt:   time.Now(),
	}

	err = app.Models.Task.Update(&task) // Pass the task pointer to the Update method
	if err != nil {
			app.errorJSON(w, errors.New("unable to update task"), http.StatusBadRequest)
			return
	}

	// log registration
	err = app.logRequest("update task", fmt.Sprintf("%s updated", task.Name))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	
	payload := jsonResponse{
			Error:    false,
			Message:  fmt.Sprintf("Updated task %s", task.Name),
			Data:     task,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) DeleteTask(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		ID int `json:"id"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	task := data.Task{ID: requestPayload.ID}

	err = task.Delete(requestPayload.ID)
	if err != nil {
		app.errorJSON(w, errors.New("unable to delete task"), http.StatusBadRequest)
		return
	}

	// log registration
	err = app.logRequest("delete task", fmt.Sprintf("%d deleted", requestPayload.ID))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:    false,
		Message:  fmt.Sprintf("deleted task %d", requestPayload.ID),
		Data:     task,
}

	app.writeJSON(w, http.StatusAccepted, payload)
}
