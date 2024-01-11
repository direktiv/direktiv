import "./i18n";

import React, { lazy } from "react";

import { createRoot } from "react-dom/client";

const App = lazy(() => import("./App"));

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

createRoot(appContainer).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
