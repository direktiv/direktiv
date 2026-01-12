import { ComponentType, Key } from "react";

import { BlockPathType } from "../PageCompiler/Block";
import { BlockType } from "../schema/blocks";
import { NoFormBlockSidePanel } from "./NoFormBlockSidePanel";
import { getBlockConfig } from "./utils/useBlockTypes";
import { isPage } from "./utils";
import { usePageEditor } from "./utils/usePageEditor";
import { usePageEditorPanel } from "./EditorPanelProvider";

export type BlockEditorAction = "create" | "edit" | "delete";

export type BlockEditFormProps<T> = {
  key: Key; // needed to ensure forms are initialized per block
  action: BlockEditorAction;
  block: T;
  path: BlockPathType;
  onSubmit: (newBlock: T) => void;
  onCancel: () => void;
};

type BlockFormProps = {
  action: BlockEditorAction;
  block: BlockType;
  path: BlockPathType;
};

export const BlockForm = ({ action, block, path }: BlockFormProps) => {
  const { addBlock, updateBlock } = usePageEditor();
  const { setPanel, panel, setDirty } = usePageEditorPanel();

  if (isPage(block)) {
    throw Error("Unexpected page object when parsing block");
  }

  if (!panel) return null;

  const handleUpdate = (newBlock: BlockType) => {
    switch (action) {
      case "create":
        addBlock(path, newBlock);
        break;
      case "edit":
        updateBlock(path, newBlock);
        break;
    }
    setDirty(false);
    setPanel(null);
  };

  const handleClose = () => {
    setDirty(false);
    setPanel(null);
  };

  // Key needed to instantiate new component per block and action
  const key = `${action}-${path.join(".")}`;

  const blockConfig = getBlockConfig(block.type);

  if (!blockConfig) {
    throw new Error("Block config must not be undefined");
  }

  const isFormBlock = !!blockConfig.formComponent;

  if (isFormBlock && block.type === blockConfig.type) {
    const FormComponent = blockConfig.formComponent as ComponentType<
      BlockEditFormProps<typeof block>
    >;

    return (
      <FormComponent
        key={key}
        action={action}
        block={block}
        path={path}
        onSubmit={handleUpdate}
        onCancel={handleClose}
      />
    );
  }

  if (!isFormBlock && block.type === blockConfig.type) {
    return (
      <NoFormBlockSidePanel
        key={key}
        action={action}
        block={block}
        path={path}
      />
    );
  }

  return (
    <div>
      Fallback form for {path} from {JSON.stringify(block)}
    </div>
  );
};
