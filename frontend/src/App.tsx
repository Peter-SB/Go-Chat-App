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
  const [displayName, setDisplayName] = useState<string>("");
  const [connected, setConnected] = useState<boolean>(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [activeUsers, setActiveUsers] = useState<string[]>([]);
  const ws = useRef<WebSocket | null>(null);

  const connectToWebSocket = () => {
    if (!displayName) {
      alert("Please enter a display name");
      return;
    }

    // Connect to WebSocket with displayName as a query parameter
    ws.current = new WebSocket(
      `ws://localhost:8080/ws?displayName=${encodeURIComponent(displayName)}`
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
      ws.current.send(message);
    }
  };

  return (
    <div className="App">
      {!connected ? (
        <div className="join-container">
          <h1 className="title">Go Chat App</h1>
          <input
            type="text"
            placeholder="Enter your display name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="input"
          />
          <button className="button" onClick={connectToWebSocket}>
            Join Chat
          </button>
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
