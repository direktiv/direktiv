import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "../PageCompiler/Block";
import { BlockList } from "../PageCompiler/Block/utils/BlockList";
import { EditorPanelLayoutProvider } from "../BlockEditor/EditorPanelProvider";
import { PageCompilerContextProvider } from "../PageCompiler/context/pageCompilerContext";
import { ParsingError } from "../PageCompiler/Block/utils/ParsingError";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type EditorLayoutProps = {
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

export const EditorLayout = ({ page, setPage }: EditorLayoutProps) => {
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
      mode="edit"
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
    </PageCompilerContextProvider>
  );
};
