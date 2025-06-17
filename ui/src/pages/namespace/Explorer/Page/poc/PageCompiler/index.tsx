import {
  PageCompilerContextProvider,
  PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { AllBlocksType } from "../schema/blocks";
import { Block } from "./Block";
import { BlockDialogProvider } from "../BlockEditor/BlockDialogProvider";
import { BlockList } from "./Block/utils/BlockList";
import { DirektivPagesSchema } from "../schema";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
import { ParsingError } from "./Block/utils/ParsingError";
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

export const PageCompiler = ({ page, setPage, mode }: PageCompilerProps) => {
  const parsedPage = DirektivPagesSchema.safeParse(page);
  const { t } = useTranslation();
  if (!parsedPage.success) {
    return (
      <ParsingError title={t("direktivPage.error.invalidSchema")}>
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </ParsingError>
    );
  }

  const onMove = (name: string, target: string, element: AllBlocksType) => {
    const newPage = page;

    newPage.blocks.splice(Number(target), 0, element);
    setPage(newPage);
  };

  return (
    <PageCompilerContextProvider setPage={setPage} page={page} mode={mode}>
      <QueryClientProvider client={queryClient}>
        <DndContext onMove={onMove}>
          <BlockDialogProvider>
            <BlockList>
              {page.blocks.map((block, index) => (
                <div key={index}>
                  <DroppableSeparator
                    visible={mode === "edit"}
                    id={String(index)}
                    position="before"
                  />
                  <Block key={index} block={block} blockPath={[index]} />
                </div>
              ))}
            </BlockList>
          </BlockDialogProvider>
        </DndContext>
      </QueryClientProvider>
    </PageCompilerContextProvider>
  );
};
