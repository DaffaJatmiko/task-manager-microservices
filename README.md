# Task Manager - Microservices

## Description

This project is an ongoing development and improvement from a previous monolithic task manager application. It has been redesigned using microservices architecture to enhance service decoupling, scalability, and fault tolerance. Each service is containerized using Docker, orchestrated using Kubernetes, and accessed via a reverse proxy (Nginx).

This project consists of a number of loosely coupled microservices, all written in Go:

- **broker-service**: Serves as a central entry point, routing requests to appropriate services (accepts JSON, communicates via gRPC, and uses RabbitMQ).
- **authentication-service**: Handles user authentication and authorization against a PostgreSQL database and generates JWT tokens (accepts JSON).
- **logger-service**: logs important events to a MongoDB database (accepts RPC, gRPC, and JSON)
- **listener-service**: consumes messages from AMQP (RabbitMQ) and initiates actions based on payload (sends via RPC)
- **mail-service**: sends email (accepts JSON)
- **task-service**: Manages tasks (CRUD operations) and requires JWT for access (accepts JSON)
- **front-end**: Provides a user-friendly web interface to interact with the services

All services (except the broker) register their access URLs with etcd and renew their leases automatically. This allows us to implement a simple service discovery system, where all service URLs are accessible with "service maps" in the Config type used to share application configuration in the broker service.

## Features

- **User Authentication and Authorization**: Secure user registration and login, with JWT token generation.
- **Task Management**: CRUD operations for tasks, secured by JWT tokens.
- **Logging**: Centralized logging for application events.
- **Email Notifications**: Sending emails using RabbitMQ for message brokering.
- **Frontend**: User-friendly web interface to interact with the services.

## Technology Stack

### Server-side Technologies

- **Programming Language**: Go
- **HTTP Router**: Chi
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Reverse Proxy**: Nginx
- **Authentication**: JWT

### Communication and Middleware

- **Message Broker**: RabbitMQ
- **gRPC**: Efficient, low-latency internal communication
- **REST**: Synchronous communication for service interactions
- **RPC**: Used for specific service interactions requiring remote procedure calls

### Database Technologies

- **Databases**: PostgreSQL, MongoDB, MySQL

### Frontend Technologies

- **Frontend**: HTML, CSS, JavaScript

### API Documentation

- **API Documentation**: Postman

## Services

1. **Authentication Service**

   - **Purpose**: Handles user registration and authentication, generating JWT tokens for authorized access.
   - **Technology**: Go, Chi, JWT
   - **Database**: PostgreSQL
   - **Communication**: HTTP REST

2. **Broker Service**

   - **Purpose**: Acts as a central hub for routing requests to the appropriate services.
   - **Technology**: Go, Chi
   - **Communication**: REST, gRPC, RabbitMQ
   - **Event Handling**: Emits and consumes events for asynchronous processing.

3. **Frontend Service**

   - **Purpose**: Provides the web interface for the application.
   - **Technology**: Javascript, HTML, CSS
   - **Communication**: HTTP REST (interacts with the broker service)

4. **Listener Service**

   - **Purpose**: Listens to events (e.g., logs, user actions) and processes them.
   - **Technology**: Go, RabbitMQ
   - **Communication**: RabbitMQ (message queue)

5. **Logger Service**

   - **Purpose**: Logs application events for monitoring and debugging.
   - **Technology**: Go, gRPC, RPC
   - **Database**: MongoDB
   - **Communication**: gRPC for internal logging, HTTP REST for log queries, RPC for specific interactions

6. **Mail Service**

   - **Purpose**: Sends email notifications.
   - **Technology**: Go, RabbitMQ
   - **Communication**: RabbitMQ for message queuing, SMTP for sending emails
   - **Templates**: HTML and plain text templates for emails

7. **Task Service**
   - **Purpose**: Manages tasks (CRUD operations). Requires JWT tokens for access, provided by the authentication service.
   - **Technology**: Go, MySQL, JWT
   - **Database**: MySQL
   - **Communication**: HTTP REST (through broker service)

## Communication Between Services

- **HTTP REST**: Synchronous communication for most service interactions.
- **gRPC**: Used for efficient, low-latency internal communication (e.g., logging).
- **RabbitMQ**: Asynchronous communication for background processing and email notifications.

## Deployment

### Prerequisites

- Docker
- Kubernetes (Minikube)
- Kubectl

### Steps to Deploy

1. **Start Minikube:**
   ```sh
   minikube start
   ```
2. **Deploy Database Containers:**

   ```sh
   docker-compose -f postgres.yml up -d
   docker-compose -f mysql.yml up -d
   ```

3. **Deploy Services to Kuberneter:**
   ```sh
   kubectl apply -f k8s/
   ```
4. **Setup Nginx Ingress Controller:**
   ```sh
   kubectl apply -f ingress.yml
   ```
5. **Access the Application:**

   - Obtain the Minikube IP:

   ```sh
   minikube ip
   ```

   - Access the application using the Minikube IP and configured ingress routes.

6. **Stopping Minikube:**
   ```sh
   minikube stop
   ```

## API Documentation

For detailed API documentation, please visit [API Documentation](https://documenter.getpostman.com/view/21784227/2sA3XJjizk).
