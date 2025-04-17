import { Block } from ".";
import { BlockWrapper } from "./utils/BlockWrapper";
import { BlocksWrapper } from "./utils/BlocksWrapper";
import { TwoColumnsType } from "../../schema/blocks/twoColumns";

export const TwoColumns = ({ leftBlocks, rightBlocks }: TwoColumnsType) => (
  <BlockWrapper>
    <BlocksWrapper horizontal>
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
    </BlocksWrapper>
  </BlockWrapper>
);
