import "./i18n";

import App from "./App";
import React from "react";
import { createRoot } from "react-dom/client";

const app = document.getElementById("root");

if (app) {
  const root = createRoot(app);
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
} else {
  throw new Error("Root element not found");
}
