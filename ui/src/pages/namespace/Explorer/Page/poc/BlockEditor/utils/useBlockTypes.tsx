import { BlockPathType } from "../../PageCompiler/Block";
import { BlockTypeConfig } from "./types";
import { blockTypes } from "./blockTypes";
import { useCallback } from "react";
import { usePage } from "../../PageCompiler/context/pageCompilerContext";

export const getBlockConfig = <T extends BlockTypeConfig["type"]>(type: T) =>
  blockTypes.find(
    (config): config is Extract<BlockTypeConfig, { type: T }> =>
      config.type === type
  );

export const useAllowedBlockTypes = () => {
  const page = usePage();

  return useCallback(
    (path: BlockPathType) =>
      blockTypes.filter((type) => type.allow(page, path)),
    [page]
  );
};
