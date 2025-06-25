import {
  PageCompilerContextProvider,
  PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlockList } from "./Block/utils/BlockList";
import { DirektivPagesSchema } from "../schema";
import { EditPanelLayoutProvider } from "../BlockEditor/EditPanelProvider";
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

  return (
    <PageCompilerContextProvider setPage={setPage} page={page} mode={mode}>
      <QueryClientProvider client={queryClient}>
        {/* rename and refactor block dialog provider */}
        <EditPanelLayoutProvider>
          <div className="flex">
            <div className="w-1/3 max-w-xs">
              {/* add EditorPane component here, with content same as dialog before */}
            </div>
            <BlockList path={[]}>
              {page.blocks.map((block, index) => (
                <Block key={index} block={block} blockPath={[index]} />
              ))}
            </BlockList>
          </div>
        </EditPanelLayoutProvider>
      </QueryClientProvider>
    </PageCompilerContextProvider>
  );
};
