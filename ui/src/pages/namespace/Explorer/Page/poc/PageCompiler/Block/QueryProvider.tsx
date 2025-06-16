import { Block, BlockPathType } from ".";
import {
  State,
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";
import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { BlockList } from "./utils/BlockList";
import { QueryProviderType } from "../../schema/blocks/queryProvider";
import { useTranslation } from "react-i18next";

type QueryProviderProps = {
  blockProps: QueryProviderType;
  blockPath: BlockPathType;
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

  const queryWithDuplicateId = queries.find(
    (query) => !!parentVariables.query[query.id]
  );

  if (queryWithDuplicateId) {
    throw new Error(
      t("direktivPage.error.duplicateId", {
        id: queryWithDuplicateId.id,
      })
    );
  }

  const queryResults: State["query"] = Object.fromEntries(
    queries.map((query, index) => [query.id, data[index]?.data])
  );

  return (
    <VariableContextProvider
      value={{
        ...parentVariables,
        query: {
          ...parentVariables.query,
          ...queryResults,
        },
      }}
    >
      <BlockList path={blockPath}>
        {blocks.map((block, index) => (
          <Block key={index} block={block} blockPath={[...blockPath, index]} />
        ))}
      </BlockList>
    </VariableContextProvider>
  );
};
