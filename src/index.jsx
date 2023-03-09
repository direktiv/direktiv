import React, { lazy } from "react";

import { createRoot } from "react-dom/client";

const app = document.getElementById("root");

const root = createRoot(app);

const App = lazy(() => {
  // lazy load is important here, if would just conditionally render one
  // or the other, the css imports of both modules would be loaded in eather case
  if (`${import.meta.env.VITE_LEGACY_DESIGN}`.toLowerCase() === "true") {
    return import("./AppLegacy");
  }
  return import("./App");
});

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
