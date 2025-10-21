import "./App.css";

import { Block } from "./pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "./pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { DirektivPagesType } from "./pages/namespace/Explorer/Page/poc/schema";
import { EditorPanelLayoutProvider } from "./pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import { PageCompilerContextProvider } from "./pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import { Toaster } from "~/design/Toast";
import { createRoot } from "react-dom/client";
import queryClient from "./util/queryClient";
import { setPage } from "./pages/namespace/Explorer/Page/poc/PageCompiler/__tests__/utils";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const page: DirektivPagesType = {
  direktiv_api: "page/v1",
  type: "page",
  blocks: [
    {
      type: "text",
      content: "Hello World",
    },
  ],
};

createRoot(appContainer).render(
  <React.StrictMode>
    <PageCompilerContextProvider setPage={setPage} page={page} mode="live">
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
  </React.StrictMode>
);
