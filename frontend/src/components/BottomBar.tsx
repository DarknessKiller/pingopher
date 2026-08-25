import React from "react";
import { HostIcon, MoonIcon, SunIcon } from "./icons";

interface BottomBarProps {
  darkMode: boolean;
  onToggleTheme: () => void;
}

const BottomBar: React.FC<BottomBarProps> = ({ darkMode, onToggleTheme }) => (
  <nav className="bottom-bar">
    <button className="bottom-bar-item active">
      <HostIcon size={20} />
      Hosts
    </button>
    <button className="bottom-bar-item" onClick={onToggleTheme}>
      {darkMode ? <SunIcon size={20} /> : <MoonIcon size={20} />}
      {darkMode ? "Light" : "Dark"}
    </button>
  </nav>
);

export default BottomBar;
