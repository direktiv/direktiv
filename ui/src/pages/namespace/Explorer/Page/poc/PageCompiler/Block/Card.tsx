import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { Card as CardDesignComponent } from "~/design/Card";
import { CardType } from "../../schema/blocks/card";

type CardProps = {
  blockProps: CardType;
  blockPath: BlockPath;
};

export const Card = ({ blockProps, blockPath }: CardProps) => (
  <CardDesignComponent className="p-5">
    <BlockList>
      {blockProps.blocks.map((block, index) => (
        <Block
          key={index}
          block={block}
          blockPath={addSegmentsToPath(blockPath, index)}
        />
      ))}
    </BlockList>
  </CardDesignComponent>
);
