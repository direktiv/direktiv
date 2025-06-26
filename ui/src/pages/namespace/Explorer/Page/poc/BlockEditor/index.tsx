import {
  useBlock,
  usePageEditor,
} from "../PageCompiler/context/pageCompilerContext";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Headline } from "./Headline";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";

export type BlockEditorAction = "create" | "edit" | "delete";

export type BlockEditFormProps<T> = {
  action: BlockEditorAction;
  block: T;
  path: BlockPathType;
  onSubmit: (newBlock: AllBlocksType) => void;
};

type BlockFormProps = {
  action: BlockEditorAction;
  path: BlockPathType;
};

export const BlockForm = ({ action, path }: BlockFormProps) => {
  const block = useBlock(path);
  const { addBlock, updateBlock } = usePageEditor();

  if (isPage(block)) {
    throw Error("Unexpected page object when parsing block");
  }

  const handleUpdate = (newBlock: AllBlocksType) => {
    switch (action) {
      case "create":
        addBlock(path, newBlock, true);
        break;
      case "edit":
        updateBlock(path, newBlock);
        break;
    }
  };

  switch (block.type) {
    case "text": {
      return (
        <Text
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
        />
      );
    }
    case "headline": {
      return (
        <Headline
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
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
