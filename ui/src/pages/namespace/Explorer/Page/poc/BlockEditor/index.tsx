import { ComponentType, Key } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { InlineBlockSidePanel } from "./InlineBlockSidePanel";
import { isPage } from "../PageCompiler/context/utils";
import { useBlockTypes } from "../PageCompiler/context/utils/useBlockTypes";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";
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
  block: AllBlocksType;
  path: BlockPathType;
};

export const BlockForm = ({ action, block, path }: BlockFormProps) => {
  const { addBlock, updateBlock } = usePageEditor();
  const { setPanel } = usePageEditorPanel();
  const { getBlockConfig } = useBlockTypes();

  if (isPage(block)) {
    throw Error("Unexpected page object when parsing block");
  }

  const handleUpdate = (newBlock: AllBlocksType) => {
    switch (action) {
      case "create":
        addBlock(path, newBlock);
        break;
      case "edit":
        updateBlock(path, newBlock);
        break;
    }
    setPanel(null);
  };

  const handleClose = () => setPanel(null);

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
      <InlineBlockSidePanel
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
