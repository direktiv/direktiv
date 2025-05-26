import {
  useBlock,
  useUpdateBlock,
} from "../PageCompiler/context/pageCompilerContext";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";

export type BlockFormProps = { path: BlockPathType; close: () => void };

export type BlockEditFormProps = {
  block: AllBlocksType;
  path: BlockPathType;
  onSave: (newBlock: AllBlocksType) => void;
};

export const BlockForm = ({ path, close }: BlockFormProps) => {
  const block = useBlock(path);
  const { updateBlock } = useUpdateBlock();

  if (Array.isArray(block)) {
    throw Error("Can not load list into block editor");
  }

  if (isPage(block)) {
    throw Error("Unexpected page object when parsing block");
  }

  const handleUpdate = (newBlock: AllBlocksType) => {
    updateBlock(path, newBlock);
    close();
  };

  switch (block.type) {
    case "text": {
      return <Text block={block} path={path} onSave={handleUpdate} />;
    }
  }

  return (
    <div>
      Block form for {path} from {JSON.stringify(block)}
    </div>
  );
};
