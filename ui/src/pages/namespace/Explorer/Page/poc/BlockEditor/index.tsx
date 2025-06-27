import { AllBlocksType } from "../schema/blocks";
import { Headline } from "./Headline";
import { QueryProvider } from "./QueryProvider";
import { Table } from "./Table";
import { Text } from "../BlockEditor/Text";
import { isPage } from "../PageCompiler/context/utils";
import { useBlockDialog } from "./BlockDialogProvider";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";

export type BlockEditorAction = "create" | "edit" | "delete";

export type BlockEditFormProps<T> = {
  block: T;
  onSubmit: (newBlock: T) => void;
};

export const BlockForm = () => {
  const { addBlock, updateBlock } = usePageEditor();
  const { setDialog, dialog } = useBlockDialog();

  if (!dialog) return null;

  const { block, path, action } = dialog;

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
    setDialog(null);
  };

  switch (block.type) {
    case "text": {
      return <Text block={block} onSubmit={handleUpdate} />;
    }
    case "headline": {
      return <Headline block={block} onSubmit={handleUpdate} />;
    }
    case "query-provider": {
      return <QueryProvider block={block} onSubmit={handleUpdate} />;
    }
    case "table": {
      return <Table block={block} onSubmit={handleUpdate} />;
    }
  }

  return (
    <div>
      Fallback form for {path} from {JSON.stringify(block)}
    </div>
  );
};
