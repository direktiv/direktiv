import { Block, BlockPathType } from ".";
import {
  VariableContextProvider,
  useGlobalVariableScope,
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
  const resolveVariableArray = useVariableArrayResolver();
  const parentVariableScope = useGlobalVariableScope();

  const variableArray = resolveVariableArray(data);

  if (parentVariableScope.loop[id]) {
    throw new Error(t("direktivPage.error.duplicateId", { id }));
  }

  if (!variableArray.success) {
    return (
      <VariableError value={data} errorCode={variableArray.error}>
        {t(`direktivPage.error.templateString.${variableArray.error}`)} (
        {variableArray.error})
      </VariableError>
    );
  }

  return (
    <BlockList path={blockPath}>
      {variableArray.data.map((item, variableIndex) => (
        <VariableContextProvider
          key={variableIndex}
          variables={{
            ...parentVariableScope,
            loop: {
              ...parentVariableScope.loop,
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
