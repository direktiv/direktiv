import { AllBlocksType, ParentBlockUnion } from "../../schema/blocks";
import { FC, PropsWithChildren, createContext, useContext } from "react";

import { BlockPath } from "../Block";
import { DirektivPagesType } from "../../schema";
import { z } from "zod";

export type State = {
  mode: "inspect" | "live";
  page: DirektivPagesType;
};

const PageCompilerContext = createContext<State | null>(null);

type PageCompilerContextProviderProps = PropsWithChildren<State>;

const PageCompilerContextProvider: FC<PageCompilerContextProviderProps> = ({
  children,
  ...value
}) => (
  <PageCompilerContext.Provider value={value}>
    {children}
  </PageCompilerContext.Provider>
);

const usePageStateContext = () => {
  const context = useContext(PageCompilerContext);
  if (!context) {
    throw new Error(
      "usePageStateContext must be used within a PageCompilerContext"
    );
  }
  return context;
};

const useMode = () => {
  const { mode } = usePageStateContext();
  return mode;
};

const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};

type Block = AllBlocksType;
type List = AllBlocksType[];
type BlockOrList = Block | List;

const isParentBlock = (
  block: AllBlocksType
): block is z.infer<typeof ParentBlockUnion> =>
  ParentBlockUnion.safeParse(block).success;

const getBlock = (list: BlockOrList, path: BlockPath): BlockOrList => {
  const result = path.reduce<BlockOrList>((acc, index) => {
    let next;

    if (Array.isArray(acc)) {
      next = acc[index];
    } else if (isParentBlock(acc)) {
      next = acc.blocks[index];
    }

    if (next) {
      return next;
    }

    throw Error(`index ${index} not found in ${JSON.stringify(acc)}`);
  }, list);
  return result;
};

const useBlock = (path: BlockPath) => {
  const page = usePage();
  return getBlock(page.blocks, path);
};

export { PageCompilerContextProvider, useMode, usePage, useBlock };
