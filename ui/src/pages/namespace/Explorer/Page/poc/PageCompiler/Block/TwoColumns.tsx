import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { BlocksWrapper } from "./utils/BlocksWrapper";
import { TwoColumnsType } from "../../schema/blocks/twoColumns";

type TwoColumnsProps = {
  blockProps: TwoColumnsType;
  blockPath: BlockPath;
};

export const TwoColumns = ({
  blockProps: { leftBlocks, rightBlocks },
  blockPath,
}: TwoColumnsProps) => (
  <BlocksWrapper horizontal>
    <BlocksWrapper>
      {leftBlocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, ["leftBlocks", index])}
        />
      ))}
    </BlocksWrapper>
    <BlocksWrapper>
      {rightBlocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, ["rightBlocks", index])}
        />
      ))}
    </BlocksWrapper>
  </BlocksWrapper>
);
