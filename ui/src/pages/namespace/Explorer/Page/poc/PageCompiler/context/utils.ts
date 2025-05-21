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

/**
 * Note: This function cannot logically return a page, but TypeScript doesn't
 * recognize that the type is narrowed down in the Array.prototype.reduce()
 * function. Thus we need to explicitly typecast the return.
 */
export const findBlock = (
  parent: AllBlocksType | DirektivPagesType,
  path: BlockPathType
): AllBlocksType => {
  if (path.length === 0) {
    throw new Error("Path must not be empty");
  }

  return path.reduce<AllBlocksType | DirektivPagesType>((acc, index) => {
    let next;

    if (isPage(acc) || isParentBlock(acc)) {
      next = acc.blocks[index] as AllBlocksType;
    }

    if (!next) {
      throw new Error(`index ${index} not found in ${JSON.stringify(acc)}`);
    }

    return next;
  }, parent) as AllBlocksType;
};

export const updateBlock = (
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

export const addBlock = (
  page: DirektivPagesType,
  block: AllBlocksType,
  path: BlockPathType
) => {
  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));
  const index = path[path.length] as number;

  if (isPage(parent) || isParentBlock(parent)) {
    const newList: AllBlocksType[] = [
      ...parent.blocks.slice(0, index - 1),
      block,
      ...parent.blocks.slice(index),
    ];

    parent.blocks = newList;
    return newPage;
  }

  throw new Error("Could not update block");
};
