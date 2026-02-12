import {
  BlockType,
  ContainerBlockType,
  containerBlockTypeList,
} from "../../schema/blocks";

const isContainerBlockType = (block: BlockType): block is ContainerBlockType =>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  containerBlockTypeList.includes(block.type as any);

export const isEmptyContainerBlock = (block: BlockType): boolean => {
  if (isContainerBlockType(block) !== true) return false;

  switch (block.type) {
    case "columns": {
      return block.blocks.every((column) => column.blocks.length === 0);
    }
    case "query-provider":
    default:
      return block.blocks.length === 0;
  }
};
