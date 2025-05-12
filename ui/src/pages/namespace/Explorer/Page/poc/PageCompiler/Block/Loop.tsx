import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { LoopType } from "../../schema/blocks/loop";
import { VariableError } from "../primitives/Variable/Error";
import { useResolveVariableArray } from "../primitives/Variable/utils/useResolveVariableArray";
import { useTranslation } from "react-i18next";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
  const { t } = useTranslation();
  const arrayVariable = useResolveVariableArray(data);

  const parentVariables = useVariables();

  if (parentVariables.loop[id]) {
    throw new Error(t("direktivPage.error.dublicateId", { id }));
  }

  if (!arrayVariable.success) {
    return (
      <VariableError value={data} errorCode={arrayVariable.error}>
        {t(`direktivPage.error.templateString.${arrayVariable.error}`)} (
        {arrayVariable.error})
      </VariableError>
    );
  }

  return (
    <BlockList>
      {arrayVariable.data.map((item, variableIndex) => (
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
          <BlockList>
            {blocks.map((block, blockIndex) => (
              <Block
                key={blockIndex}
                block={block}
                blockPath={addSegmentsToPath(blockPath, [
                  `loop#${variableIndex}`,
                  blockIndex,
                ])}
              />
            ))}
          </BlockList>
        </VariableContextProvider>
      ))}
    </BlockList>
  );
};
