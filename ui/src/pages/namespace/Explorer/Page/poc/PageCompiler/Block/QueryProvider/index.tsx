import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";

import { Block } from "..";
import { BlockWrapper } from "../utils/BlockWrapper";
import { QueryProviderType } from "../../../schema/blocks/queryProvider";

export const QueryProvider = ({ blocks, query }: QueryProviderType) => {
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
    <BlockWrapper>
      <pre>{JSON.stringify(data, null, 2)}</pre>
      {blocks.map((block, index) => (
        <Block key={index} block={block} />
      ))}
    </BlockWrapper>
  );
};
