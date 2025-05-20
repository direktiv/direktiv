import { Block, BlockPath } from ".";

import { BlockList } from "./utils/BlockList";
import { ColumnsType } from "../../schema/blocks/columns";

type ColumnsProps = {
  blockProps: ColumnsType;
  blockPath: BlockPath;
};

export const Columns = ({ blockProps, blockPath }: ColumnsProps) => (
  <BlockList horizontal>
    {blockProps.blocks.map((column, columnIndex) => (
      <BlockList key={columnIndex}>
        {column.map((block, blockIndex) => (
          <Block
            key={blockIndex}
            block={block}
            blockPath={[...blockPath, columnIndex, blockIndex]}
          />
        ))}
      </BlockList>
    ))}
  </BlockList>
);
