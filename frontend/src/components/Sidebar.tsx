import React from "react";
import { HostIcon, MoonIcon, SunIcon } from "./icons";

interface SidebarProps {
  darkMode: boolean;
  onToggleTheme: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ darkMode, onToggleTheme }) => (
  <aside className="sidebar">
    <div className="sidebar-brand">
      <h1>Pingopher</h1>
    </div>

    <nav className="sidebar-nav">
      <button className="sidebar-item active">
        <HostIcon size={18} />
        Hosts
      </button>
    </nav>

    <div className="sidebar-footer">
      <button className="theme-toggle" onClick={onToggleTheme} style={{ width: "100%" }}>
        {darkMode ? <SunIcon size={18} /> : <MoonIcon size={18} />}
        <span style={{ fontSize: "0.875rem" }}>{darkMode ? "Light Mode" : "Dark Mode"}</span>
      </button>
      <div className="sidebar-copyright">Pingopher &copy;{new Date().getFullYear()} by DarknessKiller</div>
    </div>
  </aside>
);

export default Sidebar;
