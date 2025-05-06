import { BlockPath, addSegmentsToPath } from "./utils/blockPath";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { ColumnsType } from "../../schema/blocks/columns";

type ColumnsProps = {
  blockProps: ColumnsType;
  blockPath: BlockPath;
};

export const Columns = ({ blockProps, blockPath }: ColumnsProps) => (
  <BlockList horizontal>
    {blockProps.columns.map((column, columnIndex) => (
      <BlockList key={columnIndex}>
        {column.map((block, blockIndex) => (
          <Block
            key={blockIndex}
            block={block}
            blockPath={addSegmentsToPath(blockPath, [
              "columns",
              columnIndex,
              blockIndex,
            ])}
          />
        ))}
      </BlockList>
    ))}
  </BlockList>
);
