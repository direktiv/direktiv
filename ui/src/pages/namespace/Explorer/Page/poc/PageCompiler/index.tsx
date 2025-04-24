import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import Alert from "~/design/Alert";
import { Block } from "./Block";
import { BlocksWrapper } from "./Block/utils/BlocksWrapper";
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
 * [] - add button to switv
 */

export const PageCompiler = ({ page }: PageCompilerProps) => {
  const parsedPage = DirektivPagesSchema.safeParse(page);

  if (!parsedPage.success) {
    return (
      <div className="p-4 flex flex-col gap-4">
        <Alert variant="error">
          The page schema is not valid. Please check the following error:
        </Alert>
        <pre>{JSON.stringify(parsedPage.error.issues, null, 2)}</pre>
      </div>
    );
  }

  return (
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
};
