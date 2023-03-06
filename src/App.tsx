import "./App.css";

import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";

import { getVersion } from "./api/version";

const queryClient = new QueryClient();

function App2() {
  const { isLoading, data: version } = useQuery({
    queryKey: ["version"],
    queryFn: () => getVersion({ apiKey: "password" }),
    networkMode: "always",
  });

  console.log("🚀", version);

  return <div>1{version}2</div>;
}

const App = () => (
  <QueryClientProvider client={queryClient}>
    <App2 />
  </QueryClientProvider>
);

export default App;
