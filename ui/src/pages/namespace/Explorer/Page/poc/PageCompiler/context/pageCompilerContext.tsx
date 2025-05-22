import { FC, PropsWithChildren, createContext, useContext } from "react";
import { addBlockToPage, findBlock, updateBlockInPage } from "./utils";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";

export type State = {
  mode: "inspect" | "live";
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
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

const useBlock = (path: BlockPathType) => {
  const page = usePage();
  return findBlock(page, path);
};

const useSetPage = () => {
  const { setPage } = usePageStateContext();
  return setPage;
};

export const useUpdateBlock = () => {
  const page = usePage();
  const setPage = useSetPage();
  const updateBlock = (path: BlockPathType, newBlock: AllBlocksType) => {
    const newPage = updateBlockInPage(page, path, newBlock);
    setPage(newPage);
  };

  return {
    updateBlock,
  };
};

export const useAddBlock = () => {
  const page = usePage();
  const setPage = useSetPage();
  const addBlock = (path: BlockPathType, block: AllBlocksType) => {
    const newPage = addBlockToPage(page, path, block);
    setPage(newPage);
  };

  return {
    addBlock,
  };
};

export { PageCompilerContextProvider, useMode, usePage, useSetPage, useBlock };
