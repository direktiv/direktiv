import "./i18n";

import React, { lazy } from "react";

import { createRoot } from "react-dom/client";
import env from "./config/env/";

const app = document.getElementById("root");

const root = createRoot(app);

const App = lazy(() => {
  // lazy load is important here, if would just conditionally render one
  // or the other, the css imports of both modules would be loaded in eather case
  if (env.VITE_LEGACY_DESIGN) {
    return import("./AppLegacy");
  }
  return import("./App");
});

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
