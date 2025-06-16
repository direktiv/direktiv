import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import {
  PageCompilerContextProvider,
  PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlockList } from "./Block/utils/BlockList";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
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
  const parsedPage = DirektivPagesSchema.safeParse(page);
  const { t } = useTranslation();

  if (!parsedPage.success) {
    return (
      <ParsingError title={t("direktivPage.error.invalidSchema")}>
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </ParsingError>
    );
  }

  const onMove = (
    name: string,
    target: string,
    position: "before" | "after" | undefined
  ) => {
    console.log("onmove");
  };

  return (
    <PageCompilerContextProvider setPage={setPage} page={page} mode={mode}>
      <QueryClientProvider client={queryClient}>
        <DndContext onMove={onMove}>
          <BlockList>
            {page.blocks.map((block, index) => (
              <>
                <DroppableSeparator id={String(index)} position="before" />
                <Block key={index} block={block} blockPath={[index]} />
              </>
            ))}
          </BlockList>
        </DndContext>
      </QueryClientProvider>
    </PageCompilerContextProvider>
  );
};
