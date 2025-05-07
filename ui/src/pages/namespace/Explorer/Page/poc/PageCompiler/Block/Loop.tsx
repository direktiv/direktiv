import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  VariableContextProvider,
  useVariables,
} from "../primitives/Variable/VariableContext";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { Error } from "../primitives/Variable/Error";
import { LoopType } from "../../schema/blocks/loop";
import { useResolveVariableArray } from "../primitives/Variable/utils/useResolveVariableArray";
import { useTranslation } from "react-i18next";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
  const { t } = useTranslation();
  const [variableContent, error] = useResolveVariableArray(data);

  const parentVariables = useVariables();

  if (error) {
    return (
      <Error value={data}>
        {t(`direktivPage.error.templateString.${error}`)}
      </Error>
    );
  }

  return (
    <BlockList>
      {variableContent.map((item, variableIndex) => (
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
