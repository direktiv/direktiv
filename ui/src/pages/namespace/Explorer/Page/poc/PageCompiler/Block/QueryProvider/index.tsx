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
        queryFn: () =>
          new Promise<QueryProviderType["queries"][number]>((resolve) =>
            setTimeout(() => resolve(q), 1000)
          ),
      })
    ),
  });
  const result = data.map((d) => d.data);
  return (
    <>
      <pre>{JSON.stringify(result, null, 2)}</pre>
      {blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, index)}
        />
      ))}
    </>
  );
};
