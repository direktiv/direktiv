import "../../App.css";
import "./i18n";

import { Block } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { EditorPanelLayoutProvider } from "../../pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import { PageCompilerContextProvider } from "../../pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "~/design/Toast";
import { page } from "./examplePage";
import queryClient from "../../util/queryClient";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const App = () => (
  <PageCompilerContextProvider setPage={() => {}} page={page} mode="live">
    <QueryClientProvider client={queryClient}>
      <EditorPanelLayoutProvider>
        <BlockList path={[]}>
          {page.blocks.map((block, index) => (
            <Block key={index} block={block} blockPath={[index]} />
          ))}
        </BlockList>
      </EditorPanelLayoutProvider>
    </QueryClientProvider>
    <Toaster />
  </PageCompilerContextProvider>
);

export default App;
