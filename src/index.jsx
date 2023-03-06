import App from "./App";
import AppLegacy from "./AppLegacy";

import React from "react";
import ReactDOM from "react-dom";

ReactDOM.render(
  <React.StrictMode>
    {`${import.meta.env.VITE_LEGACY_DESIGN}`.toLowerCase() === "true" ? (
      <AppLegacy />
    ) : (
      <App />
    )}
  </React.StrictMode>,
  document.getElementById("root")
);
