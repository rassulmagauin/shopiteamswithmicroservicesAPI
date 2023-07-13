# Shop items with currency Rate microservice Backend

## Overview

This project is a Go application that includes a main application and a microservice named `grpcpractice`. The `grpcpractice` microservice fetches currency exchange rates from an API provided by a European bank and provides these rates to the main application.

## Structure

- **main.go**: The entry point of the main application. It sets up the HTTP server and defines the API endpoints.

- **grpcpractice**: A microservice that fetches and provides currency exchange rates. It includes the following components:
  - **main.go**: The entry point of the microservice. It sets up a gRPC server and registers a currency server with it.
  - **server/currency.go**: Defines the `Currency` struct, which implements the `CurrencyServer` interface from the `protos` package. It includes methods for getting exchange rates and subscribing to rate updates.

- **data**: Contains data-related code, possibly for managing data models or database interactions.

- **handlers**: Contains handler functions for different routes or requests.

- **middlewares**: Contains middleware functions for handling requests and responses.

## API Endpoints

- **GET /products**: Returns a list of products. Accepts an optional `currency` query parameter to get the product prices in a specific currency.
- **GET /products/{id}**: Returns a specific product by its ID. Accepts an optional `currency` query parameter to get the product price in a specific currency.
- **PUT /products/{id}**: Updates a specific product by its ID.
- **POST /products**: Adds a new product.
- **POST /upload/{id}**: Uploads a file for a specific product by its ID.
- **GET /images/{id}**: Gets the image file for a specific product by its ID.
- **GET /token**: Gets a JWT token.

## Dependencies

The project's dependencies are managed by Go's built-in dependency management system and are listed in the `go.mod` and `go.sum` files.
