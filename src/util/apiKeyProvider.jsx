import { createContext, useContext, useState } from "react";

const ApiKeyContext = createContext();

const ApiKeyProvider = ({ children }) => (
  <ApiKeyContext.Provider value={useState(null)}>
    {children}
  </ApiKeyContext.Provider>
);

const useApiKey = () => {
  const context = useContext(ApiKeyContext);
  if (context === undefined) {
    throw new Error("useApiKey must be used within a ApiKeyProvider");
  }
  return context;
};

export { ApiKeyProvider, useApiKey };
