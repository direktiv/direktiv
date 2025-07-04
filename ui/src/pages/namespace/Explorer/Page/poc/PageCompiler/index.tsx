import { Block, BlockPathType } from "./Block";
import {
  PageCompilerContextProvider,
  PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { addBlockToPage, moveBlockWithinPage } from "./context/utils";

import { AllBlocksType } from "../schema/blocks";
import { BlockDialogProvider } from "../BlockEditor/BlockDialogProvider";
import { BlockList } from "./Block/utils/BlockList";
import { DirektivPagesSchema } from "../schema";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { ParsingError } from "./Block/utils/ParsingError";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const PageCompiler = ({ page, setPage, mode }: PageCompilerProps) => {
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

  const onMove = (
    origin: BlockPathType | null,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    const updatedPage =
      origin === null
        ? addBlockToPage(page, target, block)
        : moveBlockWithinPage(page, origin, target, block);

    setPage(updatedPage);
  };

  return (
    <PageCompilerContextProvider setPage={setPage} page={page} mode={mode}>
      <QueryClientProvider client={queryClient}>
        <DndContext onMove={onMove}>
          <BlockDialogProvider>
            <BlockList path={[]}>
              {page.blocks.map((block, index) => (
                <div key={index}>
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
