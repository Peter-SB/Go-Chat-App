import React from "react";
import "./App.css";
import "./TopBar.css";

type TopBarProps = {
  connected: boolean;
  username: string;
  onLoginClick: () => void;
  onLogoutClick: () => void;
};

const TopBar: React.FC<TopBarProps> = ({
  connected,
  username,
  onLoginClick,
  onLogoutClick,
}) => {
  return (
    <div className="top-bar">
      <div>
        <a
          href="https://www.linkedin.com/in/peter-semrau-boughton/"
          target="_blank"
          rel="noopener noreferrer"
          className="icon-button"
          title="LinkedIn"
        >
          <i className="fab fa-linkedin-in"></i>
        </a>
        <a
          href="https://github.com/Peter-SB"
          target="_blank"
          rel="noopener noreferrer"
          className="icon-button"
          title="GitHub"
        >
          <i className="fab fa-github"></i>
        </a>
      </div>
      <div className="top-bar-title"> Go Chat App </div>
      {connected && username ? (
        <div className="top-bar-content">
          <span className="welcome-text">Welcome, {username}!</span>
          <button className="top-bar-button" onClick={onLogoutClick}>
            Logout
          </button>
        </div>
      ) : (
        <div className="top-bar-content">
          <span className="welcome-text">Not logged in</span>
          <button className="top-bar-button" onClick={onLoginClick}>
            Login
          </button>
        </div>
      )}
    </div>
  );
};

export default TopBar;
