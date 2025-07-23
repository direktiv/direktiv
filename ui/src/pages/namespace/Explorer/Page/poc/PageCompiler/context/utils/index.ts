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

export const moveBlockWithinPage = (
  page: DirektivPagesType,
  originPath: BlockPathType,
  targetPath: BlockPathType,
  block: BlockType
): DirektivPagesType => {
  const originIndex = originPath[originPath.length - 1];
  const targetIndex = targetPath[targetPath.length - 1];

  if (originIndex === undefined) {
    throw new Error("Invalid path, could not extract index for origin block");
  }

  if (targetIndex === undefined) {
    throw new Error("Invalid path, could not extract index for target block");
  }

  const originParent = originPath.slice(0, -1).join("-");
  const targetParent = targetPath.slice(0, -1).join("-");

  const movingWithinSameParent = originParent === targetParent;

  const movingBefore = originIndex > targetIndex || originParent > targetParent;
  const adjustedOriginIndex =
    movingWithinSameParent && movingBefore ? originIndex + 1 : originIndex;

  const replacedLastIndexPath: BlockPathType = [
    ...originPath.slice(0, -1),
    adjustedOriginIndex,
  ];

  const pageWithAddedBlock = addBlockToPage(page, targetPath, block, false);
  const pageWithDeletedBlock = deleteBlockFromPage(
    pageWithAddedBlock,
    replacedLastIndexPath
  );
  return pageWithDeletedBlock;
};

export const decrementPath = (path: BlockPathType): BlockPathType => {
  const pathLength = path.length;
  let lastIndex = path[pathLength - 1];

  const updatedPath = lastIndex
    ? [...path.slice(0, -1), (lastIndex -= 1)]
    : path;

  return updatedPath;
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
