import { FC, ReactNode } from "react";

import { BlockPathType } from "..";
import { useBlock } from "../../context/pageCompilerContext";

type BlockProviderProps = {
  path: BlockPathType;
  children: (block: ReturnType<typeof useBlock>) => ReactNode;
};

export const BlockProvider: FC<BlockProviderProps> = ({ children, path }) => {
  const block = useBlock(path);

  return <>{children(block)}</>;
};
