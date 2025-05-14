import {
  PageCompilerContextProvider,
  State as PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlockList } from "./Block/utils/BlockList";
import { DirektivPagesSchema } from "../schema";
import { ParsingError } from "./Block/utils/ParsingError";

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

export const PageCompiler = ({ page, mode }: PageCompilerProps) => {
  const parsedPage = DirektivPagesSchema.safeParse(page);

  if (!parsedPage.success) {
    return (
      <ParsingError title="The page has an invalid configuration">
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </ParsingError>
    );
  }

  return (
    <PageCompilerContextProvider page={page} mode={mode}>
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
