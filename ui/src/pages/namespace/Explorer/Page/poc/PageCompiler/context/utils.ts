import { AllBlocksType, ParentBlockUnion } from "../../schema/blocks";
import { DirektivPagesSchema, DirektivPagesType } from "../../schema";

import { BlockPathType } from "../Block";
import { clonePage } from "../../BlockEditor/utils";
import { z } from "zod";

export const isParentBlock = (
  block: AllBlocksType
): block is z.infer<typeof ParentBlockUnion> =>
  ParentBlockUnion.safeParse(block).success;

export const isPage = (
  page: AllBlocksType | DirektivPagesType
): page is z.infer<typeof DirektivPagesSchema> =>
  DirektivPagesSchema.safeParse(page).success;

export const findBlock = (
  parent: AllBlocksType | DirektivPagesType,
  path: BlockPathType
) =>
  path.reduce<AllBlocksType | DirektivPagesType>((acc, index) => {
    let next;

    if (isPage(acc) || isParentBlock(acc)) {
      next = acc.blocks[index] as AllBlocksType;
    }

    if (!next) {
      throw new Error(`index ${index} not found in ${JSON.stringify(acc)}`);
    }

    return next;
  }, parent);

export const updateBlockInPage = (
  page: DirektivPagesType,
  path: BlockPathType,
  block: AllBlocksType
): DirektivPagesType => {
  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));
  const targetIndex = path[path.length - 1] as number;

  if (isPage(parent) || isParentBlock(parent)) {
    parent.blocks[targetIndex] = block;
    return newPage;
  }

  throw new Error("Could not update block");
};

export const addBlockToPage = (
  page: DirektivPagesType,
  path: BlockPathType,
  block: AllBlocksType
) => {
  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));
  const index = path[path.length - 1] as number;
  if (isPage(parent) || isParentBlock(parent)) {
    const newList: AllBlocksType[] = [
      ...parent.blocks.slice(0, index + 1),
      block,
      ...parent.blocks.slice(index + 1),
    ];

    parent.blocks = newList;
    return newPage;
  }

  throw new Error("Could not add block");
};

export const pathsEqual = (a: BlockPathType | null, b: BlockPathType) => {
  if (!a) {
    return false;
  }
  return a.length === b.length && a.every((val, index) => val === b[index]);
};
