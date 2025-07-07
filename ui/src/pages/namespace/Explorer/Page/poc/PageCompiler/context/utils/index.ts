import { AllBlocksType, ParentBlockUnion } from "../../../schema/blocks";
import { DirektivPagesSchema, DirektivPagesType } from "../../../schema";
import { BlockPathType } from "../../Block";
import { CardType } from "../../../schema/blocks/card";
import { ColumnsType } from "../../../schema/blocks/columns";
import { DialogType } from "../../../schema/blocks/dialog";
import { HeadlineType } from "../../../schema/blocks/headline";
import { ImageType } from "../../../schema/blocks/image";
import { LoopType } from "../../../schema/blocks/loop";
import { QueryProviderType } from "../../../schema/blocks/queryProvider";
import { TableType } from "../../../schema/blocks/table";
import { TextType } from "../../../schema/blocks/text";
import { clonePage } from "../../../BlockEditor/utils";
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
  block: AllBlocksType,
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
    const newList: AllBlocksType[] = [
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
    const newList: AllBlocksType[] = [
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
  block: AllBlocksType
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

export const idToPath = (id: string): BlockPathType => {
  const arrayOfStrings = id.split("-");
  const path = arrayOfStrings.map(Number);

  return path;
};

export const pathToId = (path: BlockPathType) => {
  const id = path.join("-");
  return id;
};

type PathOrNull = BlockPathType | null;

export const pathsEqual = (a: PathOrNull, b: PathOrNull) => {
  if (!a || !b) {
    return a === b;
  }
  return a.length === b.length && a.every((val, index) => val === b[index]);
};

type FindInBranchConfig = {
  page: DirektivPagesType;
  path: BlockPathType;
  match: (block: AllBlocksType | DirektivPagesType) => boolean;
  depth?: number;
};

export const findAncestor = ({
  page,
  path,
  match,
  depth,
}: FindInBranchConfig) => {
  if (depth !== undefined && !(depth >= 1)) {
    throw new Error("depth must be undefined or >= 1");
  }
  const limit = depth !== undefined ? path.length - depth : 0;
  for (let i = path.length - 1; i >= limit; i--) {
    const target = findBlock(page, path.slice(0, i));
    if (match(target)) {
      return true;
    }
  }
  return false;
};

export const getBlockTemplate = (type: AllBlocksType["type"]) => {
  switch (type) {
    case "headline":
      return {
        type: "headline",
        level: "h1",
        label: "",
      } satisfies HeadlineType;
    case "text":
      return {
        type: "text",
        content: "",
      } satisfies TextType;
    case "columns": {
      return {
        type: "columns",
        blocks: [
          {
            type: "column",
            blocks: [],
          },
          {
            type: "column",
            blocks: [],
          },
        ],
      } satisfies ColumnsType;
    }
    case "card":
      return {
        type: "card",
        blocks: [],
      } satisfies CardType;
    case "query-provider":
      return {
        type: "query-provider",
        blocks: [],
        queries: [],
      } satisfies QueryProviderType;
    case "table":
      return {
        type: "table",
        data: {
          type: "loop",
          id: "",
          data: "",
        },
        actions: [],
        columns: [],
      } satisfies TableType;
    case "dialog":
      return {
        type: "dialog",
        trigger: {
          type: "button",
          label: "",
        },
        blocks: [],
      } satisfies DialogType;
    case "loop":
      return {
        type: "loop",
        id: "",
        data: "",
        blocks: [],
      } satisfies LoopType;
    case "image":
      return {
        type: "image",
        src: "",
        width: 200,
        height: 200,
      } satisfies ImageType;
    default:
      throw new Error(`${type} is not implemented yet`);
  }
};
