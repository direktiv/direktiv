import { AllBlocksType } from "../schema/blocks";

export const cloneBlocks = (blocks: AllBlocksType[]): AllBlocksType[] =>
  structuredClone(blocks);
