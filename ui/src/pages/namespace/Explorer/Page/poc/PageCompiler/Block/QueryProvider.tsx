import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";
import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { QueryProviderType } from "../../schema/blocks/queryProvider";
import { useTranslation } from "react-i18next";

type QueryProviderProps = {
  blockProps: QueryProviderType;
  blockPath: BlockPath;
};

export const QueryProvider = ({
  blockProps,
  blockPath,
}: QueryProviderProps) => {
  const { blocks, queries } = blockProps;
  const { t } = useTranslation();
  const parentVariables = useVariables();
  const data = useSuspenseQueries({
    queries: queries.map((q) =>
      queryOptions({
        queryKey: [q.id],
        queryFn: async () => {
          const response = await fetch(q.endpoint);
          if (!response.ok) {
            throw new Error(
              t("direktivPage.error.queryProvider.queryFailed", {
                id: q.id,
                endpoint: q.endpoint,
                status: response.status,
              })
            );
          }
          try {
            return await response.json();
          } catch (e) {
            throw new Error(
              t("direktivPage.error.queryProvider.invalidJson", {
                id: q.id,
                endpoint: q.endpoint,
              })
            );
          }
        },
      })
    ),
  });

  let test = {};

  queries.forEach((q, i) => {
    test = { ...test, [q.id]: data[i].data };
  });

  console.log("🚀", test);

  return (
    <VariableContextProvider
      value={{
        ...parentVariables,
        query: {
          ...parentVariables.query,
          ...test,
        },
      }}
    >
      <BlockList>
        {blocks.map((block, index) => (
          <Block
            key={index}
            block={block}
            blockPath={addSegmentsToPath(blockPath, index)}
          />
        ))}
      </BlockList>
    </VariableContextProvider>
  );
};
