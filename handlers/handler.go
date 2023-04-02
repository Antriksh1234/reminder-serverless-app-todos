package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/reminder-serverless-app/reminder-serverless-app-todos/database"
	"github.com/reminder-serverless-app/reminder-serverless-app-todos/models"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case http.MethodPost:
		return CreateNewTodo(request)
	case http.MethodGet:
		return GetTodos(request)
	case http.MethodPut:
		return EditTodo(request)
	case http.MethodDelete:
		return DeleteTodo(request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       fmt.Sprintf("Unsupported HTTP method: %s", request.HTTPMethod),
		}, nil
	}
}

// Gets the todo, from the request of the API Gateway and store it in DynamoDB
func CreateNewTodo(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var todo models.Todo
	err := json.Unmarshal([]byte(request.Body), &todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
	}

	//Store the todo in DynamoDB
	if apiGatewayResponse, err := database.WriteTodoToDynamoDB(todo); err != nil {
		return apiGatewayResponse, err
	}

	responseBody, err := json.Marshal(todo)
	if err != nil {
		log.Println("CresteNewTodo(): Could not marshal todod: ", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(responseBody),
	}, nil
}

// Gets all the Todos for the user from DB
func GetTodos(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	userID := request.PathParameters["userID"]

	todos, err := database.GetTodosForUser(userID)

	//Could not get the todos from DB itself
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	responseBody, err := json.Marshal(todos)
	if err != nil {
		log.Println("GetTodos(): Could not marshal todod: ", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

// Deletes Todo from the DB
func DeleteTodo(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the ID of the Todo to delete from the request path parameters
	todoID := request.PathParameters["id"]

	//Delete from the table
	if apiGatewayResponse, err := database.DeleteTodoFromDynamoDB(todoID); err != nil {
		return apiGatewayResponse, err
	}

	//Everything went well!
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

// Updates Todo in DynamoDB
func EditTodo(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	todoID := request.PathParameters["id"]
	userID := request.PathParameters["userID"]

	// Parse the Todo data from the request body
	var todo models.Todo
	err := json.Unmarshal([]byte(request.Body), &todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
	}

	return database.UpdateTodoForUser(userID, todoID, todo)
}
