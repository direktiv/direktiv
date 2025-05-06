import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import { useEffect, useRef } from "react";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { Error } from "../Block/utils/TemplateString/Variable/Error";
import { LoopType } from "../../schema/blocks/loop";
import { useTranslation } from "react-i18next";
import { useVariableActions } from "../store/variables";
import { useVariableArray } from "./utils/TemplateString/Variable/utils/useVariableArray";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
  const isDone = useRef(false);
  const { t } = useTranslation();

  const [variableContent, error] = useVariableArray(data);

  const variableActions = useVariableActions();

  // TODO: this is a hack to get the variable to update
  useEffect(() => {
    if (variableContent && !isDone.current) {
      isDone.current = true;
      variableActions.setVariable({
        namespace: "loop",
        id,
        content: variableContent,
      });
    }
  }, [id, variableActions, variableContent]);

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
