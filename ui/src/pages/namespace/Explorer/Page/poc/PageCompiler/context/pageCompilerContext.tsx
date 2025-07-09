import {
  Dispatch,
  FC,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
} from "react";
import {
  addBlockToPage,
  deleteBlockFromPage,
  findBlock,
  moveBlockWithinPage,
  updateBlockInPage,
} from "./utils";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";

type PageCompilerMode = "edit" | "live";

export type PageCompilerProps = {
  mode: PageCompilerMode;
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
  container?: HTMLDivElement | null;
  setContainer?: Dispatch<SetStateAction<HTMLDivElement | null>>;
};

const PageCompilerContext = createContext<PageCompilerProps | null>(null);

type PageCompilerContextProviderProps = PropsWithChildren<PageCompilerProps>;

export const PageCompilerContextProvider: FC<
  PageCompilerContextProviderProps
> = ({ children, ...value }) => (
  <PageCompilerContext.Provider value={{ ...value }}>
    {children}
  </PageCompilerContext.Provider>
);

export const usePageStateContext = () => {
  const context = useContext(PageCompilerContext);
  if (!context) {
    throw new Error(
      "usePageStateContext must be used within a PageCompilerContext"
    );
  }
  return context;
};

export const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};

export const useBlock = (path: BlockPathType) => {
  const page = usePage();
  return findBlock(page, path);
};

/**
 * This hook returns variables and methods to update the page,
 * for example, creating, updating or deleting blocks.
 */
export const usePageEditor = () => {
  const page = usePage();
  const { mode, setPage } = usePageStateContext();

  const updateBlock = (path: BlockPathType, newBlock: AllBlocksType) => {
    const newPage = updateBlockInPage(page, path, newBlock);
    setPage(newPage);
  };

  const addBlock = (
    path: BlockPathType,
    block: AllBlocksType,
    after = false
  ) => {
    const newPage = addBlockToPage(page, path, block, after);
    setPage(newPage);
  };

  const deleteBlock = (path: BlockPathType) => {
    const newPage = deleteBlockFromPage(page, path);
    setPage(newPage);
  };

  const moveBlock = (
    origin: BlockPathType,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    const newPage = moveBlockWithinPage(page, origin, target, block);
    setPage(newPage);
  };

  return {
    mode,
    addBlock,
    deleteBlock,
    moveBlock,
    updateBlock,
    setPage,
  };
};
