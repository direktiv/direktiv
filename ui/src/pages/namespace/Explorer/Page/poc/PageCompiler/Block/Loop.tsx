import { Block } from ".";
import { BlockPath } from "./utils/blockPath";
import { LoopType } from "../../schema/blocks/loop";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPath;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => (
  <>
    {blockProps.blocks.map((block, index) => (
      <Block key={index} block={block} blockPath={[...blockPath, index]} />
    ))}
  </>
);
