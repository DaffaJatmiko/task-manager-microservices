package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/DaffaJatmiko/broker-service/event"
	"github.com/DaffaJatmiko/broker-service/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string `json:"action"`
	Auth AuthPayload `json:"auth,omitempty"`
	Log LogPayload `json:"log,omitempty"`
	Mail MailPayload `json:"mail,omitempty"`
	Register RegisterPayload `json:"register,omitempty"`
	AddTask AddTaskPayload `json:"add_task,omitempty"`
	GetTask GetTasksByUserIDPayload `json:"get_tasks_by_user_id,omitempty"`
	UpdateTask UpdateTaskPayload `json:"update_task,omitempty"`
	DeleteTask DeleteTaskPayload `json:"delete_task,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RegisterPayload struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type AddTaskPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int    `json:"user_id"`
}

type GetTasksByUserIDPayload struct {
	UserID int `json:"user_id"`
}

type UpdateTaskPayload struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int    `json:"user_id"`
}

type DeleteTaskPayload struct {
	ID int `json:"id"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)


}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "register":
		app.register(w, requestPayload.Register)
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logEventViaRabbit(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	// case "add_task":
	// 	app.addTask(w, requestPayload.AddTask)
	// case "get_tasks_by_user_id":
	// 	app.getTasksByUserID(w, requestPayload.GetTask)
	// case "update_task":
	// 	app.updateTask(w, requestPayload.UpdateTask)
	// case "delete_task":
	// 	app.deleteTask(w, requestPayload.DeleteTask)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) HandleTaskService(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "add_task":
		app.addTask(w, requestPayload.AddTask)
	case "get_tasks_by_user_id":
		app.getTasksByUserID(w, requestPayload.GetTask)
	case "update_task":
		app.updateTask(w, requestPayload.UpdateTask)
	case "delete_task":
		app.deleteTask(w, requestPayload.DeleteTask)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to auth microservice
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
	}

	// Log the JSON data being sent
	log.Println("Sending JSON data to auth service:", string(jsonData))

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// Log the status code from the response
	log.Println("Received status code:", response.StatusCode)

	// Read the body of the response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	// Log the response body
	log.Println("Response body:", string(body))

	// make sure we get the correct status code 
	if response.StatusCode == http.StatusUnauthorized {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("invalid creds"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		log.Println("Wrong status code 2", response.StatusCode)
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service 
	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}


func (app *Config) register(w http.ResponseWriter, r RegisterPayload) {
	jsonData, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Sending JSON data to registration service:", string(jsonData))

	request, err := http.NewRequest("POST", "http://authentication-service/register", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.Println("Received status code:", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Response body:", string(body))

	if response.StatusCode != http.StatusCreated {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("error calling registration service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Registered!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusCreated, payload)
}

func (app *Config) addTask(w http.ResponseWriter, r AddTaskPayload) {
	jsonData, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Sending JSON data to task service:", string(jsonData))

	request, err := http.NewRequest("POST", "http://task-service/tasks", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.Println("Received status code:", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Response body:", string(body))

	if response.StatusCode != http.StatusCreated {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("error calling task service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Success added task!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusCreated, payload)
}

func (app *Config) updateTask(w http.ResponseWriter, r UpdateTaskPayload) {
	jsonData, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Sending JSON data to task service:", string(jsonData))

	request, err := http.NewRequest("PUT", "http://task-service/tasks/update", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	log.Println("Received status code:", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Response body:", string(body))

	if response.StatusCode != http.StatusAccepted {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("error calling task service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Success updated task!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusCreated, payload)

}

func (app *Config) deleteTask(w http.ResponseWriter, r DeleteTaskPayload) {
	jsonData, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Sending JSON data to task service:", string(jsonData))

	request, err := http.NewRequest("DELETE", "http://task-service/tasks/delete", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.Println("Received status code:", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Response body:", string(body))

	if response.StatusCode != http.StatusAccepted {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("error calling task service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Success deleted task!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) getTasksByUserID(w http.ResponseWriter, r GetTasksByUserIDPayload) {
	jsonData, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Sending JSON data to task service:", string(jsonData))

	request, err := http.NewRequest("GET", "http://task-service/tasks/userId", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.Println("Received status code:", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body", err)
		app.errorJSON(w, err)
		return
	}

	log.Println("Response body:", string(body))

	if response.StatusCode != http.StatusOK {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("error calling task service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Success getting tasks!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusCreated, payload)
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}
	
	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Error calling log service", response.StatusCode, response)
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {

	jsonData, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		log.Println("Error marshalling data", err)
		app.errorJSON(w, err)
		return
	}

	// call the mail service
	mailServiceUrl := "http://mail-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error calling mail service", err)
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Error calling mail service", err)
		app.errorJSON(w, err)
		return
	}

	// send back json
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}