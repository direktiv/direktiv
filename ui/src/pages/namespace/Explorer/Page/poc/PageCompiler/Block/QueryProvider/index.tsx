import { BlockPath, addSegmentsToPath } from "../utils/blockPath";
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";

import { Block } from "..";
import { QueryProviderType } from "../../../schema/blocks/queryProvider";

type QueryProviderProps = {
  blockProps: QueryProviderType;
  blockPath: BlockPath;
};

export const QueryProvider = ({
  blockProps: { blocks, query },
  blockPath,
}: QueryProviderProps) => {
  const { data } = useSuspenseQuery(
    queryOptions({
      queryKey: [query.id],
      queryFn: () =>
        new Promise<QueryProviderType["query"]>((resolve) =>
          setTimeout(() => resolve(query), 1000)
        ),
    })
  );

  return (
    <>
      <pre>{JSON.stringify(data, null, 2)}</pre>
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
