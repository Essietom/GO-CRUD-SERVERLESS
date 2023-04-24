package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB

func init() {
		//connect to database
		db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		//create the table if it doesn't exist
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")

		if err != nil {
			log.Fatal(err)
		}
}

func main() {
    lambda.Start(HandleRequest)
}


func getUsers(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }
    defer rows.Close()

	users := []User{}
    for rows.Next() {
        var u User
        err := rows.Scan(&u.ID, &u.Name, &u.Email)
        if err != nil {
            log.Println(err)
            return events.APIGatewayProxyResponse{
                StatusCode: 500,
                Body:       "Internal Server Error",
            }, nil
        }
        users = append(users, u)
    }
    err = rows.Err()
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    response, err := json.Marshal(users)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       string(response),
    }, nil
}

func createUser(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    var u User
    err := json.Unmarshal([]byte(req.Body), &u)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       "Bad Request",
        }, nil
    }
	err = db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.ID)
	if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
	}
	var id int
	u.ID = id
    response, err := json.Marshal(u)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

	
    return events.APIGatewayProxyResponse{
        StatusCode: 201,
        Body:       string(response),
    }, nil
}


func updateUser(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    id, err := strconv.Atoi(req.PathParameters["id"])
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       "Bad Request",
        }, nil
    }

    var u User
    err = json.Unmarshal([]byte(req.Body), &u)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       "Bad Request",
        }, nil
    }

    stmt, err := db.Prepare("UPDATE users SET name = $1, email = $2 WHERE id = $3")
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }
    defer stmt.Close()

    result, err := stmt.Exec(u.Name, u.Email, id)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    if rowsAffected == 0 {
        return events.APIGatewayProxyResponse{
            StatusCode: 404,
            Body:       "Item Not Found",
        }, nil
    }

    u.ID = id
    response, err := json.Marshal(u)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       string(response),
    }, nil
}


func deleteUser(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    id, err := strconv.Atoi(req.PathParameters["id"])
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       "Bad Request",
        }, nil
    }

    stmt, err := db.Prepare("DELETE FROM users WHERE id = $1")
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }
    defer stmt.Close()

    result, err := stmt.Exec(id)
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Println(err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       "Internal Server Error",
        }, nil
    }

    if rowsAffected == 0 {
        return events.APIGatewayProxyResponse{
            StatusCode: 404,
            Body:       "Item Not Found",
        }, nil
    }

    return events.APIGatewayProxyResponse{
        StatusCode: 204,
        Body:       "",
    }, nil
}


func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    switch request.HTTPMethod {
		case "GET":
			return getUsers(ctx, request)
		case "POST":
			return createUser(ctx, request)
		case "PUT":
			return updateUser(ctx, request)
		case "DELETE":
			return deleteUser(ctx, request)
		default:
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "method not allowed"}, nil
		}

}

