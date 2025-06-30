import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import {
  PageCompilerContextProvider,
  PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { AllBlocksType } from "../schema/blocks";
import { Block } from "./Block";
import { BlockDialogProvider } from "../BlockEditor/BlockDialogProvider";
import { BlockList } from "./Block/utils/BlockList";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
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

  const onMove = (name: string, target: string, element: AllBlocksType) => {
    const blocks = [...page.blocks];

    const currentIndex = Number(name);
    blocks.splice(currentIndex, 1);

    const targetIndex = Number(target);
    const insertIndex =
      currentIndex < targetIndex ? targetIndex - 1 : targetIndex;

    blocks.splice(insertIndex, 0, element);

    const newPage: DirektivPagesType = {
      ...page,
      blocks,
    };

    setPage(newPage);
  };

  return (
    <PageCompilerContextProvider setPage={setPage} page={page} mode={mode}>
      <QueryClientProvider client={queryClient}>
        <DndContext onMove={onMove}>
          <BlockDialogProvider>
            <BlockList path={[]}>
              {page.blocks.map((block, index) => (
                <div key={index}>
                  {index === 0 && (
                    <DroppableSeparator
                      visible={true}
                      id={String(index)}
                      position="before"
                    />
                  )}
                  <Block key={index} block={block} blockPath={[index]} />
                  <DroppableSeparator
                    visible={true}
                    id={String(index + 1)}
                    position="after"
                  />
                </div>
              ))}
            </BlockList>
          </BlockDialogProvider>
        </DndContext>
      </QueryClientProvider>
    </PageCompilerContextProvider>
  );
};
