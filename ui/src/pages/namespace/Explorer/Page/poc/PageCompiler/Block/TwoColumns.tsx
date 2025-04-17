import { Block } from ".";
import { BlockWrapper } from "./utils/BlockWrapper";
import { BlocksWrapper } from "./utils/BlocksWrapper";
import { TwoColumnsType } from "../../schema/blocks/twoColumns";

export const TwoColumns = ({ leftBlocks, rightBlocks }: TwoColumnsType) => (
  <BlockWrapper>
    <div className="grid grid-cols-2 gap-3">
      <BlocksWrapper>
        {leftBlocks.map((block, index) => (
          <Block key={index} block={block} />
        ))}
      </BlocksWrapper>
      <BlocksWrapper>
        {rightBlocks.map((block, index) => (
          <Block key={index} block={block} />
        ))}
      </BlocksWrapper>
    </div>
  </BlockWrapper>
);
