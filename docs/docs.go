// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "xueyouchen",
            "email": "xueyou@starboardventures.io"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/gapfill": {
            "post": {
                "description": "automatic fill the gap's tipsets.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByScheduler"
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            }
        },
        "/api/v1/ping": {
            "get": {
                "description": "Healthy examination",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "Sys"
                ],
                "responses": {
                    "200": {
                        "description": "pong",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error:...",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/retry": {
            "post": {
                "description": "replay the failed tipsets.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByScheduler"
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            }
        },
        "/api/v1/task_state": {
            "post": {
                "description": "task will report tipset state with this API.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByTaskModel"
                ],
                "parameters": [
                    {
                        "description": "TipsetState",
                        "name": "TipsetState",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/core.TipsetState"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            }
        },
        "/api/v1/topic": {
            "post": {
                "description": "task group will sign in a mq topic use this API.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByTaskModel"
                ],
                "parameters": [
                    {
                        "description": "Topic",
                        "name": "Topic",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/core.Topic"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            },
            "delete": {
                "description": "delete a topic.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByTaskModel"
                ],
                "parameters": [
                    {
                        "description": "Topic",
                        "name": "Topic",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/core.Topic"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            }
        },
        "/api/v1/walk": {
            "post": {
                "description": "walk the historical DAG's tipsets.",
                "consumes": [
                    "application/json",
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/json"
                ],
                "tags": [
                    "DATA-EXTRACTION-API-Internal-V1-CallByManual"
                ],
                "parameters": [
                    {
                        "type": "boolean",
                        "name": "force",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "name": "from",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "name": "to",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "topic",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseWithRequestId"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "core.TipsetState": {
            "type": "object",
            "required": [
                "topic"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "not_found_state": {
                    "type": "integer"
                },
                "state": {
                    "type": "integer"
                },
                "tipset": {
                    "type": "integer"
                },
                "topic": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            }
        },
        "core.Topic": {
            "type": "object",
            "required": [
                "topic"
            ],
            "properties": {
                "topic": {
                    "type": "string",
                    "example": "messages/vm_messages..."
                }
            }
        },
        "utils.ResponseWithRequestId": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {},
                "message": {
                    "type": "string"
                },
                "request_id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "extractor-api.spacescope.io",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "spacescope data extraction notify backend",
	Description:      "spacescope data extraction api backend",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
