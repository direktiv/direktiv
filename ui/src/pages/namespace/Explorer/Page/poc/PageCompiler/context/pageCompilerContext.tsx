import {
  Dispatch,
  FC,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";
import {
  addBlockToPage,
  findBlock,
  pathsEqual,
  updateBlockInPage,
} from "./utils";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";

export type PageCompilerProps = {
  mode: "inspect" | "live";
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

export type PageCompilerState = PageCompilerProps & {
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
  const addBlock = (
    path: BlockPathType,
    block: AllBlocksType,
    after = false
  ) => {
    const newPage = addBlockToPage(page, path, block, after);
    setPage(newPage);
  };

  return {
    addBlock,
  };
};

const useFocus = () => {
  const { focus } = usePageStateContext();
  return { focus };
};

const useSetFocus = () => {
  const { focus, setFocus: contextSetFocus } = usePageStateContext();

  const setFocus = (path: BlockPathType) => {
    if (pathsEqual(focus, path)) {
      return contextSetFocus(null);
    }
    return contextSetFocus(path);
  };

  return setFocus;
};

export {
  PageCompilerContextProvider,
  useMode,
  usePage,
  useSetPage,
  useBlock,
  useFocus,
  useSetFocus,
};
