import {
  PageCompilerContextProvider,
  State as PageCompilerProps,
} from "./context/pageCompilerContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Block } from "./Block";
import { BlocksWrapper } from "./Block/utils/BlocksWrapper";
import { DirektivPagesSchema } from "../schema";
import { UserError } from "./Block/utils/UserError";
import { addSegmentsToPath } from "./Block/utils/blockPath";

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
      <UserError title="The page has an invalid configuration">
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </UserError>
    );
  }

  return (
    <PageCompilerContextProvider page={page} mode={mode}>
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
    </PageCompilerContextProvider>
  );
};
