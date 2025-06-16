import { Block, BlockPathType } from ".";
import {
  State,
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";

import { BlockList } from "./utils/BlockList";
import { QueryProviderType } from "../../schema/blocks/queryProvider";
import { usePageSuspenseQueries } from "../procedures/query";
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
  const data = usePageSuspenseQueries(queries);

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
      <BlockList>
        {blocks.map((block, index) => (
          <Block key={index} block={block} blockPath={[...blockPath, index]} />
        ))}
      </BlockList>
    </VariableContextProvider>
  );
};
