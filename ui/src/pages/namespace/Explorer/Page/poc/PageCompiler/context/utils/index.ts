import { BlockType, ParentBlockUnion } from "../../../schema/blocks";
import { DirektivPagesSchema, DirektivPagesType } from "../../../schema";

import { BlockPathType } from "../../Block";
import { clonePage } from "../../../BlockEditor/utils";
import { z } from "zod";

export const isParentBlock = (
  block: BlockType
): block is z.infer<typeof ParentBlockUnion> =>
  ParentBlockUnion.safeParse(block).success;

export const isPage = (
  page: BlockType | DirektivPagesType
): page is z.infer<typeof DirektivPagesSchema> =>
  DirektivPagesSchema.safeParse(page).success;

export const findBlock = (
  parent: BlockType | DirektivPagesType,
  path: BlockPathType
) =>
  path.reduce<BlockType | DirektivPagesType>((acc, index) => {
    let next;

    if (isPage(acc) || isParentBlock(acc)) {
      next = acc.blocks[index] as BlockType;
    }

    if (!next) {
      throw new Error(`index ${index} not found in ${JSON.stringify(acc)}`);
    }

    return next;
  }, parent);

export const updateBlockInPage = (
  page: DirektivPagesType,
  path: BlockPathType,
  block: BlockType
): DirektivPagesType => {
  const targetIndex = path[path.length - 1];

  if (targetIndex === undefined) {
    throw new Error("Invalid path, could not extract index for target block");
  }

  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));

  if (!(isPage(parent) || isParentBlock(parent))) {
    throw new Error("Invalid parent block");
  }
  if (!(targetIndex >= 0 && targetIndex < parent.blocks.length)) {
    throw new Error("Index for updating block out of bounds");
  }
  parent.blocks[targetIndex] = block;
  return newPage;
};

export const addBlockToPage = (
  page: DirektivPagesType,
  path: BlockPathType,
  block: BlockType,
  after = false
) => {
  let index = path[path.length - 1];

  if (index === undefined) {
    throw new Error("Invalid path, could not extract index for new block");
  }

  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));

  if (after) {
    index++;
  }

  if (isPage(parent) || isParentBlock(parent)) {
    const newList: BlockType[] = [
      ...parent.blocks.slice(0, index),
      block,
      ...parent.blocks.slice(index),
    ];

    parent.blocks = newList;
    return newPage;
  }

  throw new Error("Could not add block");
};

export const deleteBlockFromPage = (
  page: DirektivPagesType,
  path: BlockPathType
) => {
  const index = path[path.length - 1];

  if (index === undefined) {
    throw new Error("Invalid path, could not extract index for target block");
  }

  const newPage = clonePage(page);
  const parent = findBlock(newPage, path.slice(0, -1));

  if (isPage(parent) || isParentBlock(parent)) {
    const newList: BlockType[] = [
      ...parent.blocks.slice(0, index),
      ...parent.blocks.slice(index + 1),
    ];

    parent.blocks = newList;
    return newPage;
  }

  throw new Error("Could not remove block");
};

/**
 * Determines if targetPath should be updated due to deleting the item at
 * the origin path. If this affects a segment of the target path, that segment
 * is decremented in the return value
 */
export const reindexTargetPath = (
  originPath: BlockPathType,
  targetPath: BlockPathType
) => {
  if (pathsEqual(originPath, targetPath)) {
    throw new Error("origin and target paths must not be equal");
  }

  // we assume element deleted at reindexLevel shifts elements after it
  const reindexLevel = originPath.length - 1;
  const newTargetPath = [...targetPath];

  // if targetPath is shorter than reindexLevel, it won't be affected
  if (reindexLevel >= targetPath.length) return newTargetPath;

  const originIndex = originPath[reindexLevel];
  const targetIndex = targetPath[reindexLevel];

  // this should not happen thanks to early return above
  if (originIndex === undefined || targetIndex === undefined) {
    throw new Error("Unexpected mismatch between path length and reindexLevel");
  }

  const basePathsEqual = pathsEqual(
    originPath.slice(0, reindexLevel),
    targetPath.slice(0, reindexLevel)
  );

  if (basePathsEqual && targetIndex > originIndex) {
    newTargetPath[reindexLevel] = targetIndex - 1;
    return newTargetPath;
  }
  return newTargetPath;
};

export const moveBlockWithinPage = (
  page: DirektivPagesType,
  originPath: BlockPathType,
  targetPath: BlockPathType,
  block: BlockType
): DirektivPagesType => {
  if (originPath.length === 0 || targetPath.length === 0) {
    throw new Error("Paths must not be empty");
  }

  const pageWithoutOrigin = deleteBlockFromPage(page, originPath);

  const newTargetPath = reindexTargetPath(originPath, targetPath);

  const newPage = addBlockToPage(pageWithoutOrigin, newTargetPath, block);

  return newPage;
};

export const incrementPath = (path: BlockPathType): BlockPathType => {
  const pathLength = path.length;
  let lastIndex = path[pathLength - 1];

  const updatedPath =
    lastIndex !== undefined ? [...path.slice(0, -1), (lastIndex += 1)] : path;

  return updatedPath;
};

export const pathIsDescendant = (
  descendant: BlockPathType,
  ancestor: BlockPathType
): boolean => {
  if (descendant.length <= ancestor.length) return false;
  return ancestor.every((value, index) => descendant[index] === value);
};

type PathOrNull = BlockPathType | null;

export const pathsEqual = (a: PathOrNull, b: PathOrNull) => {
  if (!a || !b) {
    return a === b;
  }
  return a.length === b.length && a.every((val, index) => val === b[index]);
};

type AllPossibleBlocks = BlockType | DirektivPagesType;

type FindAncestorConfig<T extends AllPossibleBlocks["type"]> = {
  page: DirektivPagesType;
  path: BlockPathType;
  match: (
    block: AllPossibleBlocks
  ) => block is Extract<AllPossibleBlocks, { type: T }>;
  depth?: number;
};

type FindAncestorResult<T extends AllPossibleBlocks["type"]> = {
  block: Extract<AllPossibleBlocks, { type: T }>;
  path: BlockPathType;
} | null;

export const findAncestor = <T extends AllPossibleBlocks["type"]>({
  page,
  path,
  match,
  depth,
}: FindAncestorConfig<T>): FindAncestorResult<T> => {
  if (depth !== undefined && !(depth >= 1)) {
    throw new Error("depth must be undefined or >= 1");
  }
  const limit = depth !== undefined ? path.length - depth : 0;
  for (let i = path.length - 1; i >= limit; i--) {
    const targetPath = path.slice(0, i);
    const target = findBlock(page, targetPath);
    if (match(target)) {
      return {
        block: target,
        path: targetPath,
      };
    }
  }
  return null;
};
