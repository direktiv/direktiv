import { Block, BlockPathType } from ".";

import { BlockList } from "./utils/BlockList";
import { ColumnsType } from "../../schema/blocks/columns";

type ColumnsProps = {
  blockProps: ColumnsType;
  blockPath: BlockPathType;
};

export const Columns = ({ blockProps, blockPath }: ColumnsProps) => (
  <BlockList horizontal path={blockPath}>
    {blockProps.blocks.map((column, columnIndex) => (
      <BlockList key={columnIndex} path={[...blockPath, columnIndex]}>
        {column.blocks.map((block, blockIndex) => (
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
