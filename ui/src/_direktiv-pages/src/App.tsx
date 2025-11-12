import "../../App.css";
import "./i18n";

import Alert from "~/design/Alert";
import { Block } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { EditorPanelLayoutProvider } from "~/pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import { Loader2 } from "lucide-react";
import { PageCompilerContextProvider } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "~/design/Toast";
import { page } from "./examplePage";
import queryClient from "~/util/queryClient";
import { usePage } from "./api/page";
import { useTranslation } from "react-i18next";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const PageLoader = () => {
  const { data, error, isPending } = usePage(window.location.pathname);
  const { t } = useTranslation();

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Alert variant="error">
          <span className="font-bold">
            {t("direktivPage.error.genericError")}
          </span>{" "}
          {String(error)}
        </Alert>
      </div>
    );
  }

  if (isPending) {
    return <Loader2 className="mx-auto mt-10 animate-spin" />;
  }

  return (
    <PageCompilerContextProvider setPage={() => {}} page={data} mode="live">
      <EditorPanelLayoutProvider>
        <BlockList path={[]}>
          {data.blocks.map((block, index) => (
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
