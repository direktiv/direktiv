import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { InlineBlockSidePanel } from "./InlineBlockSidePanel";
import { Key } from "react";
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
  const blockTypes = useBlockTypes();

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

  // Key needed to instantiate new component per block and action
  const key = `${action}-${path.join(".")}`;

  // TODO: or should blockTypes be an object that has to implement every block type?
  const matching = blockTypes.find((type) => type.type === block.type);

  if (matching?.formComponent && block.type === matching.type) {
    const FormComponent = matching.formComponent as React.ComponentType<
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

  if (!matching?.formComponent && block.type === matching?.type) {
    return (
      <InlineBlockSidePanel
        key={key}
        action={action}
        block={block}
        path={path}
      />
    );
  }

  if (!matching) {
    return (
      <div>
        Fallback form for {path} from {JSON.stringify(block)}
      </div>
    );
  }
};
