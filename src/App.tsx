import "./App.css";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { useNamespaces } from "./api/namespaces";
import { useVersion } from "./api/version";

const queryClient = new QueryClient();

function App() {
  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } = useNamespaces();

  return (
    <div>
      <h1 className="font-bold">Version</h1>
      {isVersionLoading ? "Loading version...." : version?.api}
      <h1 className="font-bold">namespaces</h1>
      {isLoadingNamespaces
        ? "Loading namespaces"
        : namespaces?.results.map((namespace) => (
            <div key={namespace.name}>{namespace.name}</div>
          ))}
    </div>
  );
}

const AppWithQueryProvider = () => (
  <QueryClientProvider client={queryClient}>
    <App />
  </QueryClientProvider>
);

export default AppWithQueryProvider;
