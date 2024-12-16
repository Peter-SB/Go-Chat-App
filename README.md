# 📜 Project Overview

This is an **instant messaging chat app** made in part to demonstrate my abilities in full-stack development, and focusing on learning web security concepts as well as practising program architecture and scalability. It was built with a **Go** backend, a **React** frontend, and a **MySQL** database. The app is containerized using **Docker Compose** for repeatable and easy deployment.

I've taken time to implement advanced patterns like **dependency injection** for modularity, scalability, and ease of testing. I also investigated and implemented security measures such as **session management**, **CSRF tokens**, and custom **CORS middleware** for better understanding and hands-on learning. The project is complete with some example mock services and unit tests for demonstration of how to do unit tests in Go.

# 🚩Features

### Backend (Go)

- WebSockets for Real-Time chat and active user information.
- **Security Best Practices**:
  - Session Management and CSRF Protection implemented from scratch for deeper understanding.
  - Custom CORS Middleware to handle cross-origin requests.
- **Architecture**:
  - Dependency Injection for services (database and authentication), allowing easy testing with mock implementations.
  - Use of Go channels to broadcast messages to notify active users in real time.
- **Database Integration:**
  - MySQL for user authentication and message persistence.
  - Mock database implementations for unit testing.
- **Testing:**
  - Demonstration unit tests for authentication and database interactions.

### Frontend (React)

- Built with React Javascript/Typescript. Was good practice although not the focus of this project.
- Frontend communicates with the backend WebSocket API and REST endpoints.

### Database (MySQL)

- Used MySQL for practising implementing SQL despite it being overkill for this project.

### Docker

- Containerized using **Docker Compose** for easy consistent deployments.

### Tools

- **Postman**: API testing during development.
- **Git**: Version control for codebase management.

# 📸 Screenshots

<p align="center">
  <img src='docs/Screenshot-login.png'  style="width:75%;height:75%;">
</p>
<p align="center">
  <img src='docs/Screenshot-main.png'  style="width:75%;height:75%;">
</p>
<p align="center">
  <img src='docs/Screenshot-logged-out.png' style="width:75%;height:75%;">
  <p align="center"> When a user is logged out they cant connect to the websocket </p> 
</p>

# 💡 Motivation

This project was built as a learning exercise while teaching myself **Go** and exploring full-stack development and best practices for web development.

There are lots of Go frameworks such a (gin?) that handle some of the stuff implemented such as dependency injection or middleware however that aim was to practice and learn and I will be better equipped to use fraimewoks in future now I better understand the underlying mechanics.

# 📂 Project Structure (Less Important Bits Omitted)

```
chat-app/
├── backend/
│   ├── main.go          # Entry point for the Go server
│   ├── auth/            # Authentication logic
│   │   ├── auth.go      # Fnctions like Register, LoginUser and utilities for password hashing and token generation
│   │   └── auth_test.go # Unit tests for authentication functions
|   |
│   ├── broadcast/       # Handles broadcasting and notification of chat messages and active user updates
│   │   └── broadcast.go
│   ├── db/                 # Database logic and mock implementations
│   │   ├── db.go           # Functions for live MySQL database interactions (e.g., SaveMessage, GetChatHistory)
│   │   ├── db_mock.go      # Mock database implementation for testing
│   │   └── db_mock_test.go # Tests for mock database functions
|   |
│   ├── handlers/        # Request handlers for handling connections and chat history requests
│   │   └── handlers.go
│   ├── middleware/      # Custom CORS middleware to handle cross-origin requests
│   │   └── middleware.go
│   ├── models/          # Defines the data models used in the app
│   │   └── models.go
│   ├── routes/          # API route setup
│   │   └── routes.go
│   ├── services/        # Service initializations
│   │   └── services.go
│   └── utils/           # Utility functions like GetBroadcastChannel and RegisterClient
│       └── utils.go
├── frontend/
│   ├── src/
│   │   ├── App.tsx      # Main React entry point
│   │   └── Chat.tsx     # Chat component
│   │   └── TopBar.tsx   # Topbar component
│   │   ├── ....         # Other frontend bits
├── db/
│   ├── init.sql/        # Database initialisation config
├── docker-compose.yml   # Containerization configuration
└── .env                 # Example environment variables

```

# 🧑‍💻 Development Highlights

### Websockets:

I first started this project to get more hands on experiance with websockets. Initialy just for the instant messaging comunication, I later expanded this to also communicate active user updates as well.

At the moment Gorilla/websockets is defacto standard library for websockets in Go.

### **Concurrency in Go**:

This program uses concurrency by making use of Go’s Goroutines, channels, and mutex to handle tasks that can run independently and in parallel. Goroutines are lightweight threads managed by Go's runtime, allowing us to execute multiple tasks at the same time. Channels provide a way for Goroutines to communicate safely, ensuring data consistency and avoiding race conditions. Mutexes (mutual exclusions) ensure safe access to shared resources

For example, the `broadcast.StartBroadcastListener()` Goroutine listens on a shared channel to receive messages and broadcasts them to all connected clients A mutex ensures that the shared `clients` map is accessed safely:

```go
// Example Channel for broadcasting messages
var broadcast = make(chan models.Message)

// Example code from broadcast.go
// Goroutine to listen and handle messages
func StartBroadcastListener() {
	broadcast := utils.GetBroadcastChannel() //
	clients, mutex := utils.GetClients()

	for msg := range broadcast {
		messageBytes, _ := json.Marshal(msg)
		mutex.Lock() // Lock the mutex to prevent concurrent writes to the clients map

		for client := range clients {
			select {
			case client.Send <- messageBytes: // Send message to each client
			default:
				utils.DeregisterClient(client) // Remove client if unresponsive
			}
		}
		mutex.Unlock() // Unlock the mutex after processing
	}
}

// Example sending a message to the channel
func BroadcastMessage(msg models.Message) {
    broadcast <- msg // Send the message to the broadcast channel
}

// Example starting the go routine
go broadcast.StartBroadcastListener()
```

Here, `StartBroadcastListener` runs as a Goroutine and continuously listens for messages on the `broadcast` channel. When a message is received, it is sent to all connected WebSocket clients via their respective `Send` channels. This approach allows the program to handle multiple clients and messages simultaneously without blocking other tasks.

### **Session Authentication and CRFT Tokens**:

As part of this I really enjoyed learning more about session and csrf tokens, and implementing them myself from scratch. While JWT and OAuth are more modern standards, session tokens are still used a lot and learning about the security vulnerabilities introduced by those and how csrf tokens are secure against that was very interesting.

**Explanation:**

The core idea is that a session token is a way of identifying a user for a given period. This token is given to the user as a cookie when they log in and can be used to identify themselves when they make a request (such as connecting to the chat web socket or accessing their profile). benefits?

However this can introduce a vulnerability called CSRF (cross sight request forgery). Because cookies are automatically sent with requests, a malicious website could redirect an unexpecting user to make a request without the users knowing.

CSRF tokens protect against this by verifying the origin of the request. By sending a user a crsf token when they login, also as a cookie, cross-origin site policy only allowed authorised pages to access the crsf token and attach it as a customer header.

CSRF tokens are not needed everywhere though. The users with session token cookies already are automatically connected to the websocket. The browser needs to know the username to connect however and so the session-check endpoint allows the browser to check the session token validity and get the username. This endpoint however wont bother checking the CSRF token however since its a GET endpoint and not performing any actions on behalf of the user. Generally CSRF tokens are only needed for state-changing operations.

**Downsides:**

- highly distributed systems can put a strain on reading session tokens from databases if a database read is needed to check tokens for every action.
- Improper token handling (e.g. storing session tokens wrong) can cause vulnerabilities.

### Dependency Injection:

This project demonstrates Dependency Injection (DI) by using it for both the database and the auth service.

Dependency Injection is a design pattern used to achieve Inversion of Control (IoC). (IoC being a design principle where objection creation is separate from the object consuming code.) DI achieves this IoC by receiving dependencies from an external source rather than creating them internally with the objects code. DI helps improve code maintainability, testability, extendable, and flexibility by abstracting dependencies behind an interface.

In Go, rather than traditional inheritance, object orientation is achieved more through interfaces. While Go lacks class-based inheritance, polymorphism is achieved by defining interfaces and implementing their methods in Go structs. For example, we define a `DBInterface` that specifies the required methods. Any struct that implements these methods can be used interchangeably.

The `MySQLDB` struct acts as a wrapper around the actual database connection. Because it adheres to the `DBInterface`, we can swap or mock functionality without having to change the mySQL implementation.

```go
type DBInterface interface {
	SaveMessage(msg models.Message) error
	GetChatHistory() ([]models.Message, error)
	DeleteAllMessages() error
	SaveUser(username, hashedPassword string) error
	GetUserByUsername(username string) (models.User, error)
	UpdateSessionAndCSRF(userID int, sessionToken, csrfToken string) error
	ClearSession(userID int) error
	GetUserBySessionToken(sessionToken string) (models.User, error)
}

type MySQLDB struct {
	db *sql.DB
}
```

**Benefits**:

Testability: Using interfaces for DI makes it easy to replace database or auth implementations with mocks during unit testing. The auth unit tests swap out the mySQL database implementation for a MockDB.

Flexibility: Abstracting dependencies allows you to use different implementations without changing code. This is particularly useful for integrating new services like a database. By decoupling dependencies, DI reduces tight coupling between components, making the codebase easier to maintain and extend.

Separation of Concerns: DI promotes clean architecture by separating the logic of object creation from business logic, adhering to the Single Responsibility Principle (SRP).

This architecture could be further improved with the **Repository Pattern**. This would mean encapsulating our database functionality further by creating another layer of abstraction, decoupling the data access logic from other business logic. This would make testing and new service integration even easier.

### Middleware Pattern and CORS:

Because my backend was on a different port to my frontend, I had to add Cross-Origin Resource Sharing headers to my requests. To do this I implemented a Middleware pattern to sit between request and application logic to set up headers needed.

### Unit Tests:

Unit tests have been written for the auth service and the mock database. I chose to not unit test with full code coverage because I understand unit testing and this project was focused on learning and demonstaighting my abilities, not to build a full product with an actual test suit.

In Go, it is est practice is to name test files `_test.go` and put them in the same directory as the code they are testing. This is to make it easy to find the tests and supposedly encourages writing tests alongside the code. It is suggested to use separate directories for integration tests.

Within test files is is best practice to name test functions `TestXxx` where `Xxx` describes the test.

Also in Go, you can use `t.Run` to group related test cases in subtests.

### **DevOps Skills**:

Utilized Docker Compose for consistent environments and streamlined deployment.

# 🏗️ Further Expansion

- Chat paging and offset
- Repository Pattern: I was investigating other patterns such as using a repository pattern. by doing so I could increased testability of my database code and allow easy integration with out databases. However given the size of the program, and a general less is more mindset in Go, I chose to stop at dependency injection.
- WebSocket Scalability
- Rate Limiting and other security measures
- Implement CI/CD pipelines.

## 📊 Results and Insights

- Gained proficiency in **Go** for backend development.
- Practice integrate **React.js** .
- Better knowledge of **security best practices** in web development.
- Practiced **DevOps concepts** with containerized deployment using Docker.

## 🚀 How to Run

### Prerequisites

- **Docker** installed.

### Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/chat-app.git
   cd chat-app
   ```

2. Start the application using Docker Compose:

   ```bash
   Copy code
   docker-compose up --build
   ```

3. Access the app:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## 🤝 Contact

Please reach out if you have questions, always happy to talk!

- **Email**: peterboughton11@gmail.com
- **LinkedIn**: [LinkedIn](https://www.linkedin.com/in/peter-semrau-boughton/)
- **GitHub**: [GitHub](https://github.com/Peter-SB)
