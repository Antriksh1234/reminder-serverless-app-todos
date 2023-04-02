package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/reminder-serverless-app/reminder-serverless-app-todos/models"
)

// Writes the Todo to the DynamoDB table
func WriteTodoToDynamoDB(todo models.Todo) (events.APIGatewayProxyResponse, error) {

	// Initialize a new DynamoDB session
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south=1"),
	})

	if err != nil {
		log.Println("Could not create new DynamoDB session: ", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Create DynamoDB client
	svc := dynamodb.New(session)

	// Marshal the ToDo item into a DynamoDB AttributeValue
	item, err := dynamodbattribute.MarshalMap(todo)
	if err != nil {
		log.Println("Could not marshal todo into a dynamodbattributevalue: ", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Create the DynamoDB PutItemInput object
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("Todos"),
	}

	// Put the ToDo item in DynamoDB
	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Could not put item in DynamoDB: ", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

// Deletes the Todo from the DynamoDB table
func DeleteTodoFromDynamoDB(todoID string) (events.APIGatewayProxyResponse, error) {
	// Create a new DynamoDB client
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Delete the Todo item from the DynamoDB table
	_, err := svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Todos"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(todoID),
			},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func GetTodosForUser(userId string) ([]models.Todo, error) {
	// Create a new DynamoDB client
	svc := dynamodb.New(session.New())

	// Create a QueryInput object for the request
	input := &dynamodb.QueryInput{
		TableName:              aws.String("Todod"),
		KeyConditionExpression: aws.String("userId = :userId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userId": {
				S: aws.String(userId),
			},
		},
	}

	// Call the Query API
	result, err := svc.Query(input)
	if err != nil {
		log.Println("Error while querying: ", err.Error())
		return nil, err
	}

	// Unmarshal the items into an array of Todo objects
	todos := []models.Todo{}
	for _, item := range result.Items {
		var todo models.Todo
		err = dynamodbattribute.UnmarshalMap(item, &todo)
		if err != nil {
			log.Println("Error while Unmarshalling: ", err.Error())
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// Updates the todo for a user in DynamoDB
func UpdateTodoForUser(userID string, todoID string, todo models.Todo) (events.APIGatewayProxyResponse, error) {

	// Create an update expression to update the fields
	updateExpr := "SET title = :t, description = :d, completed = :c"
	exprVals := map[string]interface{}{
		":t": todo.Title,
		":d": todo.Description,
		":c": todo.Completed,
	}

	// Create a DynamoDB client
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	})
	svc := dynamodb.New(sess)

	// Define the input for the update item operation
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Todos"),
		Key: map[string]*dynamodb.AttributeValue{
			"userID": {
				S: aws.String(userID),
			},
			"todoID": {
				S: aws.String(todoID),
			},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: convertToAttributeValueMap(exprVals),
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	// Update the item in the table
	_, err := svc.UpdateItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	responseBody, err := json.Marshal(todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

// Helper function to convert a map to a DynamoDB AttributeValue map
func convertToAttributeValueMap(m map[string]interface{}) map[string]*dynamodb.AttributeValue {
	result := make(map[string]*dynamodb.AttributeValue)
	for k, v := range m {
		result[k] = &dynamodb.AttributeValue{
			S: aws.String(fmt.Sprintf("%v", v)),
		}
	}
	return result
}
