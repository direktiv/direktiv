import { FC, PropsWithChildren, createContext, useContext } from "react";

import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";
import { findBlock } from "./utils";

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

export { PageCompilerContextProvider, useMode, usePage, useSetPage, useBlock };
