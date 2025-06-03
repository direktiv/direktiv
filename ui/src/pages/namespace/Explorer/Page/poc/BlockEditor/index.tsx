import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { DirektivPagesType } from "../schema";
import { Headline } from "./Headline";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";

export type BlockEditorAction = "create" | "edit";

/**
 * Please use this exported type for components that edit blocks. Omit the
 * generic blocks type from BlockEditFormProps, and set it to the specific
 * block type.
 */
export type BlockEditFormProps = {
  action: BlockEditorAction;
  block: AllBlocksType;
  path: BlockPathType;
  onSubmit: (newBlock: AllBlocksType) => void;
};

export type BlockFormProps = {
  block: AllBlocksType | DirektivPagesType;
  action: BlockEditorAction;
  path: BlockPathType;
};

export const BlockForm = ({ action, path, block }: BlockFormProps) => {
  const { addBlock, updateBlock } = usePageEditor();

  if (Array.isArray(block)) {
    throw Error("Cannot load list into block editor");
  }

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
