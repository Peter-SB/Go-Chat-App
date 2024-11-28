import React, { useState } from "react";
import "./Chat.css";

type Message = {
  sender: string;
  content: string;
  timestamp: string;
};

type ChatProps = {
  messages: Message[];
  sendMessage: (message: string) => void;
};

const Chat: React.FC<ChatProps> = ({ messages, sendMessage }) => {
  const [message, setMessage] = useState<string>("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() !== "") {
      sendMessage(message);
      setMessage("");
    }
  };

  return (
    <div className="chat-container">
      <div className="chatbox">
        {messages.length === 0 ? ( // Check if messages array is empty
          <div className="no-messages">
            No messages yet. Start the conversation!
          </div>
        ) : (
          messages.map((msg, index) => (
            <div key={index} className="message">
              <span className="message-sender">{msg.sender}</span>
              <span className="message-content">{msg.content}</span>
              <span className="message-timestamp">
                {new Date(msg.timestamp).toLocaleTimeString()}
              </span>
            </div>
          ))
        )}
      </div>
      <form className="message-form" onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Type a message..."
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          className="message-input"
        />
        <button type="submit" className="message-button">
          Send
        </button>
      </form>
    </div>
  );
};

export default Chat;
