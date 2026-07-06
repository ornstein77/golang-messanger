// Package docs contains the generated OpenAPI (Swagger) specification.
//
// This file is normally produced by `swag init` (see `make swagger`). A
// hand-maintained baseline is committed so the project builds and serves the
// Swagger UI out of the box; regenerate it whenever annotations change.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support"
        },
        "license": {
            "name": "MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Verifies credentials and returns a bearer session token.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["auth"],
                "summary": "Log in",
                "parameters": [
                    {
                        "description": "Login data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": { "$ref": "#/definitions/loginRequest" }
                    }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/sessionEnvelope" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/errorEnvelope" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "security": [{ "BearerAuth": [] }],
                "description": "Revokes the current session token.",
                "produces": ["application/json"],
                "tags": ["auth"],
                "summary": "Log out",
                "responses": {
                    "204": { "description": "No Content" },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Creates an account with a nickname, password and phone number.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["auth"],
                "summary": "Register a new user",
                "parameters": [
                    {
                        "description": "Registration data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": { "$ref": "#/definitions/registerRequest" }
                    }
                ],
                "responses": {
                    "201": { "description": "Created", "schema": { "$ref": "#/definitions/userEnvelope" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/errorEnvelope" } },
                    "409": { "description": "Conflict", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        },
        "/messages": {
            "get": {
                "security": [{ "BearerAuth": [] }],
                "description": "Returns the messages exchanged with another user, oldest first.",
                "produces": ["application/json"],
                "tags": ["messages"],
                "summary": "Conversation history",
                "parameters": [
                    { "type": "string", "description": "Other user's UUID", "name": "with", "in": "query", "required": true },
                    { "type": "integer", "description": "Max results (default 50, max 200)", "name": "limit", "in": "query" },
                    { "type": "integer", "description": "Pagination offset", "name": "offset", "in": "query" }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/messageListEnvelope" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/errorEnvelope" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            },
            "post": {
                "security": [{ "BearerAuth": [] }],
                "description": "Sends a direct message to another user.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["messages"],
                "summary": "Send a message",
                "parameters": [
                    {
                        "description": "Message payload",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": { "$ref": "#/definitions/sendRequest" }
                    }
                ],
                "responses": {
                    "201": { "description": "Created", "schema": { "$ref": "#/definitions/messageEnvelope" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/errorEnvelope" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } },
                    "404": { "description": "Not Found", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        },
        "/users": {
            "get": {
                "security": [{ "BearerAuth": [] }],
                "description": "Lists other registered users so you can start a conversation.",
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "List users",
                "parameters": [
                    { "type": "integer", "description": "Max results (default 50, max 200)", "name": "limit", "in": "query" },
                    { "type": "integer", "description": "Pagination offset", "name": "offset", "in": "query" }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/userListEnvelope" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        },
        "/users/me": {
            "get": {
                "security": [{ "BearerAuth": [] }],
                "description": "Returns the authenticated user's profile.",
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Current profile",
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/userEnvelope" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/errorEnvelope" } }
                }
            }
        }
    },
    "definitions": {
        "registerRequest": {
            "type": "object",
            "required": ["nickname", "password", "phone"],
            "properties": {
                "nickname": { "type": "string", "example": "john_doe" },
                "password": { "type": "string", "example": "s3cr3t-pass" },
                "phone": { "type": "string", "example": "+14155552671" }
            }
        },
        "loginRequest": {
            "type": "object",
            "required": ["nickname", "password"],
            "properties": {
                "nickname": { "type": "string", "example": "john_doe" },
                "password": { "type": "string", "example": "s3cr3t-pass" }
            }
        },
        "sendRequest": {
            "type": "object",
            "required": ["body", "recipient_id"],
            "properties": {
                "recipient_id": { "type": "string", "example": "6f9619ff-8b86-d011-b42d-00c04fc964ff" },
                "body": { "type": "string", "example": "Hey there!" }
            }
        },
        "userResponse": {
            "type": "object",
            "properties": {
                "id": { "type": "string" },
                "nickname": { "type": "string" },
                "phone": { "type": "string" },
                "created_at": { "type": "string" }
            }
        },
        "sessionResponse": {
            "type": "object",
            "properties": {
                "token": { "type": "string" },
                "expires_at": { "type": "string" },
                "user": { "$ref": "#/definitions/userResponse" }
            }
        },
        "messageResponse": {
            "type": "object",
            "properties": {
                "id": { "type": "string" },
                "sender_id": { "type": "string" },
                "recipient_id": { "type": "string" },
                "body": { "type": "string" },
                "created_at": { "type": "string" }
            }
        },
        "errorBody": {
            "type": "object",
            "properties": {
                "code": { "type": "string" },
                "message": { "type": "string" },
                "fields": { "type": "object", "additionalProperties": { "type": "string" } }
            }
        },
        "errorEnvelope": {
            "type": "object",
            "properties": { "error": { "$ref": "#/definitions/errorBody" } }
        },
        "userEnvelope": {
            "type": "object",
            "properties": { "data": { "$ref": "#/definitions/userResponse" } }
        },
        "userListEnvelope": {
            "type": "object",
            "properties": { "data": { "type": "array", "items": { "$ref": "#/definitions/userResponse" } } }
        },
        "sessionEnvelope": {
            "type": "object",
            "properties": { "data": { "$ref": "#/definitions/sessionResponse" } }
        },
        "messageEnvelope": {
            "type": "object",
            "properties": { "data": { "$ref": "#/definitions/messageResponse" } }
        },
        "messageListEnvelope": {
            "type": "object",
            "properties": { "data": { "type": "array", "items": { "$ref": "#/definitions/messageResponse" } } }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer\" followed by a space and the session token.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger metadata consumed by swaggo/http-swagger.
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Messenger API",
	Description:      "A small messenger backend.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
