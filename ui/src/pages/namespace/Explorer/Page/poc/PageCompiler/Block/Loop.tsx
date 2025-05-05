import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { LoopType } from "../../schema/blocks/loop";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, variable } = blockProps;

  return (
    <>
      looping over {variable}
      {blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, index)}
        />
      ))}
    </>
  );
};
