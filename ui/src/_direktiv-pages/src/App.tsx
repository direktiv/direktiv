import "../../App.css";
import "./i18n";

import { Block } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { DirektivPagesSchema } from "~/pages/namespace/Explorer/Page/poc/schema";
import { EditorPanelLayoutProvider } from "~/pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import ErrorMessage from "./Error";
import { Loader2 } from "lucide-react";
import { PageCompilerContextProvider } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "~/design/Toast";
import queryClient from "~/util/queryClient";
import { usePage } from "./api/page";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const PageLoader = () => {
  const { data, error, isPending } = usePage(window.location.pathname);

  if (isPending) {
    return <Loader2 className="mx-auto mt-10 animate-spin" />;
  }

  if (error) {
    return <ErrorMessage error={String(error)} />;
  }

  const page = DirektivPagesSchema.safeParse(data);
  if (!page.success) {
    return <ErrorMessage error={page.error.message} />;
  }

  return (
    <PageCompilerContextProvider
      setPage={() => {}}
      page={page.data}
      mode="live"
    >
      <EditorPanelLayoutProvider>
        <BlockList path={[]}>
          {page.data.blocks.map((block, index) => (
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
