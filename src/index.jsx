import App from "./App";
import AppLegacy from "./AppLegacy";
import React from "react";
import { createRoot } from "react-dom/client";
const app = document.getElementById("root");
const root = createRoot(app);

root.render(
  <React.StrictMode>
    {`${import.meta.env.VITE_LEGACY_DESIGN}`.toLowerCase() === "true" ? (
      <AppLegacy />
    ) : (
      <App />
    )}
  </React.StrictMode>
);
