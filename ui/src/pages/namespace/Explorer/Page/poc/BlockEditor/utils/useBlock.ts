import { BlockPathType } from "../../PageCompiler/Block";
import { findBlock } from ".";
import { usePage } from "../../PageCompiler/context/pageCompilerContext";

export const useBlock = (path: BlockPathType) => {
  const page = usePage();
  return findBlock(page, path);
};
