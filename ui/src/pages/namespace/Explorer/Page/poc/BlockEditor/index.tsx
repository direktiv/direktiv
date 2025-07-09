import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Dialog } from "./Dialog";
import { Form } from "./Form";
import { Headline } from "./Headline";
import { Image } from "./Image";
import { InlineBlockSidePanel } from "./InlineBlockSidePanel";
import { Key } from "react";
import { Loop } from "./Loop";
import { QueryProvider } from "./QueryProvider";
import { Table } from "./Table";
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

  switch (block.type) {
    case "text": {
      return (
        <Text
          key={key}
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
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "image": {
      return (
        <Image
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "query-provider": {
      return (
        <QueryProvider
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "card":
    case "columns": {
      return (
        <InlineBlockSidePanel
          key={key}
          action={action}
          block={block}
          path={path}
        />
      );
    }

    case "table": {
      return (
        <Table
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "dialog": {
      return (
        <Dialog
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "loop": {
      return (
        <Loop
          key={key}
          action={action}
          block={block}
          path={path}
          onSubmit={handleUpdate}
          onCancel={handleClose}
        />
      );
    }
    case "form": {
      return (
        <Form
          key={key}
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
