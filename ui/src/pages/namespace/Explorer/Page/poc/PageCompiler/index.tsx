import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import {
  PageCompilerContextProvider,
  PageCompilerMode,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlockList } from "./Block/utils/BlockList";
import { EditorPanelLayoutProvider } from "../BlockEditor/EditorPanelProvider";
import { ParsingError } from "./Block/utils/ParsingError";
import { Toaster } from "~/design/Toast";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type PageCompilerProps = {
  mode: PageCompilerMode;
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

export const PageCompiler = ({ page, setPage, mode }: PageCompilerProps) => {
  const [scrollPos, setScrollPos] = useState(0);

  const [queryClient] = useState(
    new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          networkMode: "always", // the default networkMode sometimes assumes that the client is offline
        },
        mutations: {
          retry: false,
          networkMode: "always", // the default networkMode sometimes assumes that the client is offline
        },
      },
    })
  );

  const parsedPage = DirektivPagesSchema.safeParse(page);
  const { t } = useTranslation();
  if (!parsedPage.success) {
    return (
      <ParsingError title={t("direktivPage.error.invalidSchema")}>
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </ParsingError>
    );
  }

  return (
    <PageCompilerContextProvider
      setPage={setPage}
      page={page}
      mode={mode}
      scrollPos={scrollPos}
      setScrollPos={setScrollPos}
    >
      <QueryClientProvider client={queryClient}>
        <EditorPanelLayoutProvider>
          <BlockList path={[]}>
            {page.blocks.map((block, index) => (
              <Block key={index} block={block} blockPath={[index]} />
            ))}
          </BlockList>
        </EditorPanelLayoutProvider>
      </QueryClientProvider>
      {/* When embedded in Direktiv to be used as a preview, there is already a toaster provider on the page. */}
      {mode === "live" && <Toaster />}
    </PageCompilerContextProvider>
  );
};
