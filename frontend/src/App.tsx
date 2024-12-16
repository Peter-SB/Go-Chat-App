import React, { useState, useEffect, useRef } from "react";
import Chat from "./Chat";
import TopBar from "./TopBar";
import "./App.css";

type Message = {
  sender: string;
  content: string;
  timestamp: string;
};

const App: React.FC = () => {
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [csrfToken, setCsrfToken] = useState<string>("");
  const [connected, setConnected] = useState<boolean>(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [activeUsers, setActiveUsers] = useState<string[]>([]);
  const [showLoginPopup, setShowLoginPopup] = useState<boolean>(true);

  const ws = useRef<WebSocket | null>(null);
  const ipAddress = window.location.hostname;

  // Automatically connect the user in if they already have valid session tokens
  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch(`http://${ipAddress}:8080/session-check`, {
          method: "GET",
          credentials: "include", // This ensures cookies are included
        });

        if (response.ok) {
          const data = await response.json();
          console.log("Session is valid. User:", data.username);
          setUsername(data.username);

          const tokenSet = await setCSRFTokenFromCookies();
          if (tokenSet) {
            const csrfCookie = document.cookie
              .split("; ")
              .find((row) => row.startsWith("csrf_token="));
            const token = csrfCookie ? csrfCookie.split("=")[1] : "";
            connectToWebSocket(data.username, token);
            setShowLoginPopup(false);
          }
        } else {
          console.log("Session is invalid. Showing login popup.");
          setShowLoginPopup(true);
        }
      } catch (error) {
        console.error("Error checking session:", error);
        setShowLoginPopup(true);
      }
    };

    checkSession();
  }, []);

  // Fetch message history from the /history endpoint
  const fetchMessageHistory = async () => {
    try {
      console.log("Fetching chat history");
      const response = await fetch(`http://${ipAddress}:8080/history`);
      if (response.ok) {
        const history: Message[] = await response.json();
        if (history === null) {
          console.log("No chat history available.");
          setMessages([]); // Set an empty array if the response is null to stop render errors
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
        const tokenSet = await setCSRFTokenFromCookies();
        if (tokenSet) {
          const csrfCookie = document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="));
          const token = csrfCookie ? csrfCookie.split("=")[1] : "";
          connectToWebSocket(username, token);
          setShowLoginPopup(false);
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

  const handleLogout = async () => {
    try {
      const response = await fetch(`http://${ipAddress}:8080/logout`, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "X-CSRF-Token": csrfToken,
        },
        body: new URLSearchParams({ username, password }),
        credentials: "include",
      });

      if (response.ok) {
        // Close WebSocket if open
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
          ws.current.close();
        }
        setConnected(false);
        setUsername("");
        setPassword("");
        setCsrfToken("");
        setActiveUsers([]);
        setShowLoginPopup(true);
      } else {
        const errorText = await response.text();
        alert(`Logout failed: ${errorText}`);
      }
    } catch (error) {
      console.error("Logout error:", error);
      alert("An error occurred during logout.");
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

  const connectToWebSocket = (username: string, csrfToken: string) => {
    if (!username) {
      alert("Please enter a display name");
      return;
    }

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
    if (!connected) {
      alert("You must be logged in to send messages.");
      return;
    }
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const formattedMessage = JSON.stringify({
        sender: username,
        content: message,
        timestamp: new Date().toISOString(),
      });
      ws.current.send(formattedMessage);
    }
  };

  const handleLoginSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    handleLogin();
  };

  return (
    <div className="App">
      <TopBar
        connected={connected}
        username={username}
        onLoginClick={() => setShowLoginPopup(true)}
        onLogoutClick={handleLogout}
      />

      <div className="chat-layout">
        <Chat
          messages={messages}
          sendMessage={sendMessage}
          canSend={connected}
        />
        <div className="user-list">
          <h2>Active Users</h2>
          {!connected ? (
            <p className="greyed-out">
              Only logged in users can see active chat members
            </p>
          ) : (
            <ul>
              {activeUsers.map((user, index) => (
                <li key={index}>
                  <span className="user-dot online"></span>
                  {user}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Login Popup Overlay */}
      {!connected && showLoginPopup && (
        <div className="login-overlay">
          <div className="login-popup">
            <button
              className="close-button"
              onClick={() => setShowLoginPopup(false)}
            >
              X
            </button>
            <h1 className="title">Go Chat App</h1>
            <form onSubmit={handleLoginSubmit}>
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
                <button type="submit" className="button">
                  Login
                </button>
                <button
                  type="button"
                  className="button"
                  onClick={handleRegister}
                >
                  Register
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default App;
