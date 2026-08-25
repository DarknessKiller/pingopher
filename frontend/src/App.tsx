import React, { useEffect, useState } from "react";
import Sidebar from "./components/Sidebar";
import BottomBar from "./components/BottomBar";
import { ToastProvider } from "./components/Toast";
import HostList from "./components/HostList";

const App: React.FC = () => {
  const [darkMode, setDarkMode] = useState(() => {
    if (typeof window !== "undefined" && window.matchMedia) {
      return window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    return false;
  });

  useEffect(() => {
    document.documentElement.dataset.theme = darkMode ? "dark" : "light";
  }, [darkMode]);

  useEffect(() => {
    if (!window.matchMedia) return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = (e: MediaQueryListEvent) => setDarkMode(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return (
    <ToastProvider>
      <div className="app">
        <Sidebar darkMode={darkMode} onToggleTheme={() => setDarkMode((d) => !d)} />
        <main className="content">
          <HostList />
          <div className="content-footer">Pingopher &copy;{new Date().getFullYear()} by DarknessKiller</div>
        </main>
        <BottomBar darkMode={darkMode} onToggleTheme={() => setDarkMode((d) => !d)} />
      </div>
    </ToastProvider>
  );
};

export default App;
