import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { TwoColumnsType } from "../../schema/blocks/twoColumns";

type TwoColumnsProps = {
  blockProps: TwoColumnsType;
  blockPath: BlockPath;
};

export const TwoColumns = ({
  blockProps: { leftBlocks, rightBlocks },
  blockPath,
}: TwoColumnsProps) => (
  <BlockList horizontal>
    <BlockList>
      {leftBlocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, ["leftBlocks", index])}
        />
      ))}
    </BlockList>
    <BlockList>
      {rightBlocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, ["rightBlocks", index])}
        />
      ))}
    </BlockList>
  </BlockList>
);
