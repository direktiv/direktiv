import { BlockPath, addSegmentsToPath } from "../utils/blockPath";
import { LoopIdContextProvider, useLoopIndex } from "./LoopIdContext";

import { Block } from "..";
import { BlockList } from "../utils/BlockList";
import { Error } from "../utils/TemplateString/Variable/Error";
import { LoopType } from "../../../schema/blocks/loop";
import { useInitLoopVariable } from "./useInitLoopVariable";
import { useTranslation } from "react-i18next";
import { useVariableArray } from "../utils/TemplateString/Variable/utils/useVariableArray";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
  const { t } = useTranslation();
  const [variableContent, error] = useVariableArray(data);
  const parentLoopIndex = useLoopIndex();

  useInitLoopVariable(id, variableContent);

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
        <LoopIdContextProvider
          key={variableIndex}
          value={{
            ...parentLoopIndex,
            [id]: variableIndex,
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
        </LoopIdContextProvider>
      ))}
    </BlockList>
  );
};
