import {
  Dispatch,
  FC,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";
import { addBlockToPage, pathsEqual, updateBlockInPage } from "./utils";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";

type PageCompilerMode = "edit" | "live";

export type PageCompilerProps = {
  mode: PageCompilerMode;
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

type PageCompilerState = PageCompilerProps & {
  focus: BlockPathType | null;
  setFocus: Dispatch<SetStateAction<BlockPathType | null>>;
};

const PageCompilerContext = createContext<PageCompilerState | null>(null);

type PageCompilerContextProviderProps = PropsWithChildren<PageCompilerProps>;

const PageCompilerContextProvider: FC<PageCompilerContextProviderProps> = ({
  children,
  ...value
}) => {
  const [focus, setFocus] = useState<BlockPathType | null>(null);
  return (
    <PageCompilerContext.Provider value={{ ...value, focus, setFocus }}>
      {children}
    </PageCompilerContext.Provider>
  );
};

const usePageStateContext = () => {
  const context = useContext(PageCompilerContext);
  if (!context) {
    throw new Error(
      "usePageStateContext must be used within a PageCompilerContext"
    );
  }
  return context;
};

const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};

// Todo: Currently not used. Remove it if we don't need it later.

// const useBlock = (path: BlockPathType) => {
//   const page = usePage();
//   return findBlock(page, path);
// };

export const usePageEditor = () => {
  const page = usePage();
  const {
    focus,
    mode,
    setFocus: contextSetFocus,
    setPage,
  } = usePageStateContext();

  const setFocus = (path: BlockPathType) => {
    if (focus && pathsEqual(focus, path)) {
      return !!contextSetFocus && contextSetFocus(null);
    }
    return !!contextSetFocus && contextSetFocus(path);
  };

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

  return {
    focus,
    mode,
    setFocus,
    addBlock,
    updateBlock,
    setPage,
  };
};

export { PageCompilerContextProvider };
