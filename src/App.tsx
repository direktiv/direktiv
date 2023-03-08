import "./App.css";

import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";

import { getNamespaces } from "./api/namespaces";
import { useVersion } from "./api/version";

const queryClient = new QueryClient();

function App() {
  const { data: version } = useVersion();
  const { data: namespaces } = useQuery({
    queryKey: ["namespaces"],
    queryFn: () =>
      getNamespaces({
        apiKey: "password",
        params: undefined,
      }),
    networkMode: "always",
  });

  return (
    <div>
      {version?.api}
      <h1>namespaces</h1>
      {namespaces?.results.map((namespace) => (
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
