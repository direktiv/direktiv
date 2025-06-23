import { AllBlocksType, inlineBlockTypes } from "../../schema/blocks";
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
  deleteBlockFromPage,
  getBlockTemplate,
  pathsEqual,
  updateBlockInPage,
} from "./utils";

import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";
import { useBlockDialog } from "../../BlockEditor/BlockDialogProvider";

export type PageCompilerMode = "edit" | "live";

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

export const PageCompilerContextProvider: FC<
  PageCompilerContextProviderProps
> = ({ children, ...value }) => {
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

export const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};

// Todo: Currently not used. Remove it if we don't need it later.

// const useBlock = (path: BlockPathType) => {
//   const page = usePage();
//   return findBlock(page, path);
// };

/**
 * This hook returns variables and methods to update the page,
 * for example, creating, updating or deleting blocks.
 */
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
    contextSetFocus(null);
  };

  const deleteBlock = (path: BlockPathType) => {
    const newPage = deleteBlockFromPage(page, path);
    setPage(newPage);
    contextSetFocus(null);
  };

  return {
    focus,
    mode,
    setFocus,
    addBlock,
    deleteBlock,
    updateBlock,
    setPage,
  };
};

/**
 * This hook returns createBlock(), which opens the editor dialog for
 * blocks such as text blocks, or just adds an inline block to the page
 * (e.g., cards or columns, where no dialog is required).
 */
export const useCreateBlock = () => {
  const { addBlock } = usePageEditor();
  const { setDialog } = useBlockDialog();

  const createBlock = (type: AllBlocksType["type"], path: BlockPathType) => {
    if (inlineBlockTypes.includes(type)) {
      return addBlock(path, getBlockTemplate(type), true);
    }
    setDialog({
      action: "create",
      block: getBlockTemplate(type),
      path,
    });
  };

  return { createBlock };
};
