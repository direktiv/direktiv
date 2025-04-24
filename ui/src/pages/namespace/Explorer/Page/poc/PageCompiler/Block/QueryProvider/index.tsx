import { BlockPath, addSegmentsToPath } from "../utils/blockPath";
import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { Block } from "..";
import { QueryProviderType } from "../../../schema/blocks/queryProvider";

type QueryProviderProps = {
  blockProps: QueryProviderType;
  blockPath: BlockPath;
};

export const QueryProvider = ({
  blockProps: { blocks, queries },
  blockPath,
}: QueryProviderProps) => {
  const data = useSuspenseQueries({
    queries: queries.map((q) =>
      queryOptions({
        queryKey: [q.id],
        queryFn: async () => {
          const response = await fetch(q.endpoint);
          if (!response.ok) {
            throw new Error(
              `Error in query with the id  ${q.id}. Failed to fetch test data, API responded with ${response.status}`
            );
          }
          try {
            return await response.json();
          } catch (e) {
            throw new Error(
              `Error in query with the id  ${q.id}. Data is not valid JSON`
            );
          }
        },
      })
    ),
  });

  const result = data.map((d) => d.data);

  return (
    <>
      {blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, index)}
        />
      ))}
      <pre>{JSON.stringify(result, null, 2)}</pre>
    </>
  );
};
