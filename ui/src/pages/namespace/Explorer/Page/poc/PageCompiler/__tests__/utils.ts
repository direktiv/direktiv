import { AllBlocksType } from "../../schema/blocks";
import { DirektivPagesType } from "../../schema";

export const createDirektivPage = (
  blocks: AllBlocksType[]
): DirektivPagesType => ({
  direktiv_api: "pages/v1",
  blocks,
});
