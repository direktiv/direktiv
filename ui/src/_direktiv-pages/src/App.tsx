import "../../App.css";
import "./i18n";

import {
  QueryClientProvider,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";

import { Block } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { EditorPanelLayoutProvider } from "../../pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import { Loader2 } from "lucide-react";
import { PageCompilerContextProvider } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { Toaster } from "~/design/Toast";
import { page } from "./examplePage";
import queryClient from "../../util/queryClient";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const PageLoader = () => {
  const queryOpts = queryOptions({
    queryKey: ["page"],
    queryFn: async () => {
      const res = await fetch(window.location.pathname + "/page.json");
      if (!res.ok) throw new Error("Failed to fetch page");
      return res.json();
    },
  });

  const { data, isLoading, error } = useQuery(queryOpts);

  if (error) {
    return <div>Error loading page</div>;
  }

  if (isLoading) {
    return <Loader2 className="mx-auto mt-10 size-8 animate-spin" />;
  }

  return (
    <PageCompilerContextProvider
      setPage={() => {}}
      page={data as typeof page}
      mode="live"
    >
      <EditorPanelLayoutProvider>
        <BlockList path={[]}>
          {(data as typeof page).blocks.map((block, index) => (
            <Block key={index} block={block} blockPath={[index]} />
          ))}
        </BlockList>
      </EditorPanelLayoutProvider>
    </PageCompilerContextProvider>
  );
};

const App = () => (
  <QueryClientProvider client={queryClient}>
    <PageLoader />
    <Toaster />
  </QueryClientProvider>
);

export default App;
