import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { Block } from ".";
import { QueryProviderType } from "../../schema/blocks/queryProvider";

type QueryProviderProps = {
  blockProps: QueryProviderType;
  blockPath: BlockPath;
};

export const QueryProvider = ({
  blockProps,
  blockPath,
}: QueryProviderProps) => {
  const { blocks, queries } = blockProps;
  useSuspenseQueries({
    queries: queries.map((q) =>
      queryOptions({
        queryKey: [q.id],
        queryFn: async () => {
          const response = await fetch(q.endpoint);
          if (!response.ok) {
            throw new Error(
              `Error in query with id ${q.id}. GET ${q.endpoint} responded with ${response.status}`
            );
          }
          try {
            return await response.json();
          } catch (e) {
            throw new Error(
              `Error in query with id ${q.id}. GET ${q.endpoint} returned invalid JSON`
            );
          }
        },
      })
    ),
  });

  return (
    <>
      {blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, ["blocks", index])}
        />
      ))}
    </>
  );
};
