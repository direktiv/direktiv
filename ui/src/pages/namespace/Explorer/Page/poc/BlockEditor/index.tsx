import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Headline } from "./Headline";
import { Key } from "react";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";
import { usePageEditorPanel } from "./EditorPanelProvider";

export type BlockEditorAction = "create" | "edit" | "delete";

export type BlockEditFormProps<T> = {
  key: Key; // needed to ensure forms are initialized per block
  action: BlockEditorAction;
  block: T;
  path: BlockPathType;
  onSubmit: (newBlock: AllBlocksType) => void;
  onCancel: () => void;
};

type BlockFormProps = {
  action: BlockEditorAction;
  block: AllBlocksType;
  path: BlockPathType;
};

export const BlockForm = ({ action, block, path }: BlockFormProps) => {
  const { addBlock, updateBlock } = usePageEditor();
  const { setPanel } = usePageEditorPanel();

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
    setPanel(null);
  };

  const handleClose = () => setPanel(null);

  switch (block.type) {
    case "text": {
      return (
        <Text
          key={path.join()}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "headline": {
      return (
        <Headline
          key={path.join()}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
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
