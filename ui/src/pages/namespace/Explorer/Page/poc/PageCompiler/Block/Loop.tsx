import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { Error } from "../Block/utils/TemplateString/Variable/Error";
import { LoopType } from "../../schema/blocks/loop";
import { useTranslation } from "react-i18next";
import { useVariableArray } from "./utils/TemplateString/Variable/utils/useVariableArray";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, variable } = blockProps;
  const { t } = useTranslation();

  const [variableContent, error] = useVariableArray(variable);

  if (error) {
    return (
      <Error value={variable}>
        {t(`direktivPage.error.templateString.${error}`)}
      </Error>
    );
  }

  return (
    <BlockList>
      {variableContent.map((item, variableIndex) => (
        <BlockList key={variableIndex}>
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
      ))}
    </BlockList>
  );
};
