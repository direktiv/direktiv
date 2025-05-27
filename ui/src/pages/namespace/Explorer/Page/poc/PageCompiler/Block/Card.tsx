import { Block, BlockPathType } from ".";

import { BlockList } from "./utils/BlockList";
import { Card as CardDesignComponent } from "~/design/Card";
import { CardType } from "../../schema/blocks/card";

type CardProps = {
  blockProps: CardType;
  blockPath: BlockPathType;
};

export const Card = ({ blockProps, blockPath }: CardProps) => (
  <CardDesignComponent className="p-5">
    <BlockList>
      {blockProps.blocks.map((block, index) => (
        <Block key={index} block={block} blockPath={[...blockPath, index]} />
      ))}
    </BlockList>
  </CardDesignComponent>
);
