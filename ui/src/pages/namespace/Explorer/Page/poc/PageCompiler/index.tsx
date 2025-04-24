import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlocksWrapper } from "./Block/utils/BlocksWrapper";
import { DirektivPagesType } from "../schema";
import { addSegmentsToPath } from "./Block/utils/blockPath";

type PageCompilerProps = {
  page: DirektivPagesType;
};

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

/**
 *
 * TODO:
 * [] add context provider
 * [] - mode is either "preview" | "live"
 * [] - containing the json in a variable
 */

export const PageCompiler = ({ page }: PageCompilerProps) => (
  <QueryClientProvider client={queryClient}>
    <BlocksWrapper>
      {page.blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath("blocks", index)}
        />
      ))}
    </BlocksWrapper>
  </QueryClientProvider>
);
