import { AllBlocksType } from "../schema/blocks";
import { DirektivPagesType } from "../schema";

export const cloneBlocks = (blocks: AllBlocksType[]): AllBlocksType[] =>
  structuredClone(blocks);

export const clonePage = (page: DirektivPagesType): DirektivPagesType =>
  structuredClone(page);
