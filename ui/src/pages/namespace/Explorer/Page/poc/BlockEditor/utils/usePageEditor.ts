import {
  addBlockToPage,
  deleteBlockFromPage,
  moveBlockWithinPage,
  updateBlockInPage,
} from "./updatePage";
import {
  usePage,
  usePageStateContext,
} from "../../PageCompiler/context/pageCompilerContext";

import { BlockPathType } from "../../PageCompiler/Block";
import { BlockType } from "../../schema/blocks";

/**
 * This hook returns variables and methods to update the page,
 * for example, creating, updating or deleting blocks.
 */
export const usePageEditor = () => {
  const page = usePage();
  const { mode, setPage } = usePageStateContext();

  const updateBlock = (path: BlockPathType, newBlock: BlockType) => {
    const newPage = updateBlockInPage(page, path, newBlock);
    setPage(newPage);
  };

  const addBlock = (path: BlockPathType, block: BlockType, after = false) => {
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
    block: BlockType
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
