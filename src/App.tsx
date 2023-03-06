import "./App.css";

import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";

import { getVersion } from "./api/version";

const queryClient = new QueryClient();

function App() {
  const { data: version } = useQuery({
    queryKey: ["version"],
    queryFn: () => getVersion({ apiKey: "password" }),
    // networkMode: "always",
  });

  return <div>{version?.api}</div>;
}

const AppWithQueryProvider = () => (
  <QueryClientProvider client={queryClient}>
    <App />
  </QueryClientProvider>
);

export default AppWithQueryProvider;
