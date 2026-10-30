import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";
import App from "./App";

document.addEventListener(
  "wheel",
  (e) => {
    const el = e.target as HTMLElement;
    if (el instanceof HTMLInputElement && el.type === "number") el.blur();
  },
  { passive: true }
);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
