import { Block, BlockPathType } from "./Block";
import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import {
  PageCompilerContextProvider,
  State as PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { BlockList } from "./Block/utils/BlockList";
import { ParsingError } from "./Block/utils/ParsingError";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const queryClient = new QueryClient({
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
});

export const PageCompiler = ({
  page: initialPage,
  mode,
}: PageCompilerProps) => {
  const [page, setPage] = useState<DirektivPagesType>(initialPage ?? "");
  const [focus, setFocus] = useState<BlockPathType | null>(null);
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
      setFocus={setFocus}
      page={page}
      focus={focus}
      mode={mode}
    >
      <QueryClientProvider client={queryClient}>
        <BlockList>
          {page.blocks.map((block, index) => (
            <Block key={index} block={block} blockPath={[index]} />
          ))}
        </BlockList>
      </QueryClientProvider>
    </PageCompilerContextProvider>
  );
};
