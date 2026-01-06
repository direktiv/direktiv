import "../../App.css";
import "./i18n";

import App from "./App";
import React from "react";
import { createRoot } from "react-dom/client";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

createRoot(appContainer).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
