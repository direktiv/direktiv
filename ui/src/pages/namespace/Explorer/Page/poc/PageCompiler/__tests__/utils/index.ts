import { BlockType } from "../../../schema/blocks";
import { DirektivPagesType } from "../../../schema";

export const createDirektivPage = (blocks: BlockType[]): DirektivPagesType => ({
  direktiv_api: "page/v1",
  type: "page",
  blocks,
});
