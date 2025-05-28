import {
  useAddBlock,
  useBlock,
  useUpdateBlock,
} from "../PageCompiler/context/pageCompilerContext";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Headline } from "./Headline";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";

export type BlockEditorAction = "create" | "edit";

export type BlockFormProps = {
  action: BlockEditorAction;
  path: BlockPathType;
  close: () => void;
};

export type BlockEditFormProps = {
  block: AllBlocksType;
  path: BlockPathType;
  onSave: (newBlock: AllBlocksType) => void;
};

export const BlockForm = ({ action, path, close }: BlockFormProps) => {
  const block = useBlock(path); // TODO: only needed when editing!
  const { updateBlock } = useUpdateBlock();
  const { addBlock } = useAddBlock();

  if (Array.isArray(block)) {
    throw Error("Can not load list into block editor");
  }

  if (isPage(block)) {
    throw Error("Unexpected page object when parsing block");
  }

  const handleUpdate = (newBlock: AllBlocksType) => {
    switch (action) {
      case "create":
        addBlock(path, newBlock); // TODO: add block after (requires merge first)
        break;
      case "edit":
        updateBlock(path, newBlock);
        break;
    }
    close();
  };

  switch (block.type) {
    case "text": {
      return (
        <Text action={action} block={block} path={path} onSave={handleUpdate} />
      );
    }
    case "headline": {
      return (
        <Headline
          action={action}
          block={block}
          path={path}
          onSave={handleUpdate}
        />
      );
    }
  }

  return (
    <div>
      Fallback form for {path} from {JSON.stringify(block)}
    </div>
  );
};
