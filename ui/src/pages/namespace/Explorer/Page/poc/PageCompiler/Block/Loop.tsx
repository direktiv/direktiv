import { Block, BlockPathType } from ".";
import {
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";

import { BlockList } from "./utils/BlockList";
import { LoopType } from "../../schema/blocks/loop";
import { VariableError } from "../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../primitives/Variable/utils/useVariableArrayResolver";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPathType;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
  const { t } = useTranslation();
  const resolvedVariableArray = useVariableArrayResolver()(data);

  const parentVariables = useVariables();

  if (parentVariables.loop[id]) {
    throw new Error(t("direktivPage.error.duplicateId", { id }));
  }

  if (!resolvedVariableArray.success) {
    return (
      <VariableError value={data} errorCode={resolvedVariableArray.error}>
        {t(`direktivPage.error.templateString.${resolvedVariableArray.error}`)}{" "}
        ({resolvedVariableArray.error})
      </VariableError>
    );
  }

  return (
    <BlockList path={blockPath}>
      {resolvedVariableArray.data.map((item, variableIndex) => (
        <VariableContextProvider
          key={variableIndex}
          value={{
            ...parentVariables,
            loop: {
              ...parentVariables.loop,
              [id]: item,
            },
          }}
        >
          <BlockList path={blockPath}>
            {blocks.map((block, blockIndex) => {
              const path = [...blockPath, blockIndex];
              return (
                <Block key={path.join(".")} block={block} blockPath={path} />
              );
            })}
          </BlockList>
        </VariableContextProvider>
      ))}
    </BlockList>
  );
};
