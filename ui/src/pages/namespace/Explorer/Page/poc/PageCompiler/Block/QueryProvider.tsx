import { Block, BlockPathType } from ".";
import {
  ContextVariables,
  VariableContextProvider,
  useVariablesContext,
} from "../primitives/Variable/VariableContext";

import { BlockList } from "page-blocklist";
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
  const parentVariables = useVariablesContext();
  const data = usePageSuspenseQueries(queries);

  const queryWithDuplicateId = queries.find(
    (query) => !!parentVariables.query?.[query.id]
  );

  if (queryWithDuplicateId) {
    throw new Error(
      t("direktivPage.error.duplicateId", {
        id: queryWithDuplicateId.id,
      })
    );
  }

  const queryResults: ContextVariables["query"] = Object.fromEntries(
    queries.map((query, index) => [query.id, data[index]?.data])
  );

  return (
    <VariableContextProvider
      variables={{
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
