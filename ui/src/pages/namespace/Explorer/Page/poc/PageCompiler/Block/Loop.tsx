import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { LoopType } from "../../schema/blocks/loop";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps: { blocks }, blockPath }: LoopProps) => (
  <>
    {/* TODO: add iteration logic */}
    {blocks.map((block, index) => (
      <Block
        key={index}
        block={block}
        blockPath={addSegmentsToPath(blockPath, index)}
      />
    ))}
  </>
);
