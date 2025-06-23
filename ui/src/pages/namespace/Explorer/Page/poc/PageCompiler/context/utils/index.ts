import { AllBlocksType, ParentBlockUnion } from "../../../schema/blocks";
import { DirektivPagesSchema, DirektivPagesType } from "../../../schema";
import { BlockPathType } from "../../Block";
import { CardType } from "../../../schema/blocks/card";
import { ColumnsType } from "../../../schema/blocks/columns";
import { HeadlineType } from "../../../schema/blocks/headline";
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

type PathOrNull = BlockPathType | null;

export const pathsEqual = (a: PathOrNull, b: PathOrNull) => {
  if (!a || !b) {
    return a === b;
  }
  return a.length === b.length && a.every((val, index) => val === b[index]);
};

type ParseAncestorsConfig = {
  page: DirektivPagesType;
  path: BlockPathType;
  fn: (block: AllBlocksType | DirektivPagesType) => boolean;
  depth?: number;
};

export const parseAncestors = ({
  page,
  path,
  fn,
  depth,
}: ParseAncestorsConfig) => {
  const limit = depth ? path.length - depth - 1 : 0;
  for (let i = path.length - 1; i > limit; i--) {
    const ancestor = findBlock(page, path.slice(0, i));
    if (fn(ancestor)) {
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
    default:
      return { type: "text", content: "" } satisfies TextType;
  }
};
