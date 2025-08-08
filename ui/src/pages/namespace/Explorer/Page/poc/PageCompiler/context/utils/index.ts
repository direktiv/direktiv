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

export const findParentBlock = (
  page: DirektivPagesType,
  path: BlockPathType
): BlockType | DirektivPagesType => {
  let current: BlockType | DirektivPagesType = page;

  for (let i = 0; i < path.length - 1; i++) {
    const index = path[i] ?? 0;

    if (!isPage(current) && !isParentBlock(current)) {
      throw new Error(
        `Block at path [${path.slice(0, i).join(",")}] is not a parent block`
      );
    }

    let next: BlockType | undefined = current.blocks[index];
    if (!next) {
      throw new Error(
        `No block at index ${index} in path [${path.slice(0, i).join(",")}]`
      );
    }
    next = next as BlockType;
    current = next;
  }

  return current;
};

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

export const isMovingBefore = (
  originPath: BlockPathType,
  targetPath: BlockPathType
) => {
  const len = Math.min(originPath.length, targetPath.length);

  for (let i = 0; i < len; i++) {
    const origin = originPath[i];
    const target = targetPath[i];

    if (origin === undefined || target === undefined) {
      throw new Error("Paths Indices must not be undefined");
    }

    if (origin >= target) return true;
    if (origin < target) return false;
  }
  throw new Error("Paths should never be equal");
};

export const firstDifferentIndex = (
  originPath: BlockPathType,
  targetPath: BlockPathType
) => {
  const len = Math.min(originPath.length, targetPath.length);
  for (let i = 0; i < len; i++) {
    if (originPath[i] !== targetPath[i]) return i;
  }
  return originPath.length !== targetPath.length ? len : null;
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

  const sameParent =
    originPath.length === targetPath.length &&
    originPath.slice(0, -1).every((v, i) => v === targetPath[i]);

  const targetOnRootLevel = targetPath.length === 1;

  const index = firstDifferentIndex(originPath, targetPath) ?? 0;

  const targetBeforeOrigin = isMovingBefore(originPath, targetPath);

  const adjustedOriginPath: BlockPathType = [...originPath];

  if (targetBeforeOrigin && targetOnRootLevel && adjustedOriginPath[0]) {
    adjustedOriginPath[0] += 1;
  }

  if (
    !targetOnRootLevel &&
    targetBeforeOrigin &&
    sameParent &&
    adjustedOriginPath[index]
  ) {
    adjustedOriginPath[index] += 1;
  }

  const pageWithAddedBlock = addBlockToPage(page, targetPath, block, false);

  const pageWithDeletedBlock = deleteBlockFromPage(
    pageWithAddedBlock,
    adjustedOriginPath
  );

  return pageWithDeletedBlock;
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
