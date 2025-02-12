// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/download/{path}": {
            "get": {
                "description": "Download a file from the specified path",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "download"
                ],
                "summary": "Download a file",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Path to the file to download",
                        "name": "path",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized - Invalid signature",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden - Invalid signature",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "File not found",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "410": {
                        "description": "Gone - Token expired",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/image/{path}": {
            "get": {
                "description": "Get an image with optional transformations applied",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "image/jpeg",
                    "image/png",
                    "image/webp"
                ],
                "tags": [
                    "image"
                ],
                "summary": "Process and transform an image",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Path to the image file",
                        "name": "path",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Output image width in pixels",
                        "name": "w",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Output image height in pixels",
                        "name": "h",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "clip",
                            "crop",
                            "fill"
                        ],
                        "type": "string",
                        "description": "Resize mode: clip, crop, fill",
                        "name": "fit",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "jpg",
                            "jpeg",
                            "png",
                            "webp"
                        ],
                        "type": "string",
                        "description": "Output format: jpg, jpeg, png, webp",
                        "name": "fm",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Compression quality (1-100)",
                        "name": "q",
                        "in": "query"
                    },
                    {
                        "type": "number",
                        "description": "Device pixel ratio (1-3)",
                        "name": "dpr",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Gaussian blur intensity (0-100)",
                        "name": "blur",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Force download instead of display",
                        "name": "dl",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized - Invalid signature",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden - Invalid signature",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Image not found",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "410": {
                        "description": "Gone - Token expired",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/list/{path}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get a list of files and directories at the specified path",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "list"
                ],
                "summary": "List contents of a directory",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Path to list contents from",
                        "name": "path",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of files and directories",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/utils.RcloneFile"
                            }
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized - Invalid or missing API key",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Path not found",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/utils.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "utils.ErrorResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "details": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                }
            }
        },
        "utils.RcloneFile": {
            "type": "object",
            "properties": {
                "IsDir": {
                    "type": "boolean"
                },
                "MimeType": {
                    "type": "string"
                },
                "ModTime": {
                    "type": "string"
                },
                "Name": {
                    "type": "string"
                },
                "Path": {
                    "type": "string"
                },
                "Size": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "description": "Type \"Bearer\" followed by a space and API key.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "/v2",
	Schemes:          []string{},
	Title:            "shuto API",
	Description:      "API for processing and transforming images",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
