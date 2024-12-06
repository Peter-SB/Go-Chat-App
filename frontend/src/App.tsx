import React, { useState, useEffect, useRef } from "react";
import Chat from "./Chat";
import "./App.css";

type Message = {
  sender: string;
  content: string;
  timestamp: string;
};

type ActiveUsersMessage = {
  type: string;
  users: string[];
};

const App: React.FC = () => {
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [csrfToken, setCsrfToken] = useState<string>("");
  const [connected, setConnected] = useState<boolean>(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [activeUsers, setActiveUsers] = useState<string[]>([]);
  const ws = useRef<WebSocket | null>(null);
  const ipAddress = window.location.hostname;

  // Fetch message history from the /history endpoint
  const fetchMessageHistory = async () => {
    try {
      console.log("Fetching chat history");
      const response = await fetch(`http://${ipAddress}:8080/history`);
      if (response.ok) {
        const history: Message[] = await response.json();
        if (history === null) {
          console.log("No chat history available.");
          setMessages([]); // Set an empty array if the response is null
        } else {
          console.log(history);
          setMessages(history);
        }
      } else {
        console.error("Failed to fetch message history");
      }
    } catch (error) {
      console.error("Error fetching message history:", error);
    }
  };

  useEffect(() => {
    fetchMessageHistory();
  }, []);

  const handleRegister = async () => {
    try {
      const response = await fetch(`http://${ipAddress}:8080/register`, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ username, password }),
        credentials: "include", // Include cookies for session handling
      });

      if (response.ok) {
        alert("Registration successful! Logging you in...");
        handleLogin(); // Automatically log in after registration
      } else {
        const errorText = await response.text();
        alert(`Registration failed: ${errorText}`);
      }
    } catch (error) {
      console.error("Registration error:", error);
      alert("An error occurred during registration.");
    }
  };

  const handleLogin = async () => {
    console.log("logging in");
    try {
      const response = await fetch(`http://${ipAddress}:8080/login`, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ username, password }),
        credentials: "include",
      });

      if (response.ok) {
        alert("Login successful!");
        const tokenSet = await setCSRFTokenFromCookies();
        if (tokenSet) {
          const csrfCookie = document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="));
          const token = csrfCookie ? csrfCookie.split("=")[1] : "";
          connectToWebSocket(token);
        } else {
          alert(
            "CSRF token could not be retrieved. WebSocket connection aborted."
          );
        }
      } else {
        const errorText = await response.text();
        alert(`Login failed: ${errorText}`);
      }
    } catch (error) {
      console.error("Login error:", error);
      alert("An error occurred during login.");
    }
  };

  const setCSRFTokenFromCookies = async (): Promise<boolean> => {
    return new Promise((resolve) => {
      const csrfCookie = document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="));
      if (csrfCookie) {
        const token = csrfCookie.split("=")[1];
        console.log("CSRF Token:", token);
        setCsrfToken(token);
        resolve(true);
      } else {
        console.error("CSRF token not found in cookies.");
        resolve(false);
      }
    });
  };

  const connectToWebSocket = (csrfToken: string) => {
    if (!username) {
      alert("Please enter a display name");
      return;
    }

    console.log("CSRF Token:", csrfToken);
    if (!csrfToken) {
      alert("Missing CSRF token. Please log in again.");
      return;
    }

    ws.current = new WebSocket(
      `ws://${ipAddress}:8080/ws?csrf_token=${csrfToken}`
    );

    ws.current.onmessage = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      if (data.type === "activeUsers") {
        // Update active users list
        setActiveUsers(data.users);
      } else {
        const message: Message = {
          sender: data.sender,
          content: data.content,
          timestamp: data.timestamp,
        };
        setMessages((prevMessages) => [...prevMessages, message]);
      }
    };

    ws.current.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.current.onclose = () => {
      console.log("WebSocket connection closed");
      setConnected(false);
    };

    setConnected(true);
  };

  const sendMessage = (message: string) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const formattedMessage = JSON.stringify({
        sender: username,
        content: message,
        timestamp: new Date().toISOString(),
      });
      ws.current.send(formattedMessage);
    }
  };

  return (
    <div className="App">
      {!connected ? (
        <div className="join-container">
          <h1 className="title">Go Chat App</h1>
          <input
            type="text"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="input"
          />
          <input
            type="password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="input"
          />
          <div className="button-container">
            <button className="button" onClick={handleLogin}>
              Login
            </button>
            <button className="button" onClick={handleRegister}>
              Register
            </button>
          </div>
        </div>
      ) : (
        <div className="chat-layout">
          <Chat messages={messages} sendMessage={sendMessage} />
          <div className="user-list">
            <h2>Active Users</h2>
            <ul>
              {activeUsers.map((user, index) => (
                <li key={index}>
                  <span className="user-dot online"></span>
                  {user}
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  );
};

export default App;
