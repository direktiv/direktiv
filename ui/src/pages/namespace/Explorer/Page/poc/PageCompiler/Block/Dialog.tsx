import { Block, BlockPathType } from ".";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import { LocalDialog, LocalDialogContent } from "~/components/LocalDialog";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import { DialogType } from "../../schema/blocks/dialog";
import { useLocalDialogContainer } from "~/components/LocalDialog/container";
import { usePageEditor } from "../context/pageCompilerContext";
import { usePageEditorPanel } from "../../BlockEditor/EditorPanelProvider";

type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPathType;
  onOpenChange?: (open: boolean) => void;
};

const DialogBaseComponent = ({
  blockProps,
  blockPath,
  onOpenChange,
}: DialogProps) => {
  const { container } = useLocalDialogContainer();
  const { blocks, trigger } = blockProps;

  return (
    <LocalDialog onOpenChange={onOpenChange}>
      <DialogTrigger>
        <Button data-local-dialog-trigger blockProps={trigger} />
      </DialogTrigger>

      <LocalDialogContent container={container}>
        <DialogXClose />
        <BlockList path={blockPath}>
          {blocks.map((block, index) => (
            <Block
              key={index}
              block={block}
              blockPath={[...blockPath, index]}
            />
          ))}
        </BlockList>
      </LocalDialogContent>
    </LocalDialog>
  );
};

const EditModeDialog = (props: DialogProps) => {
  const { setPanel } = usePageEditorPanel();

  return (
    <DialogBaseComponent
      {...props}
      onOpenChange={() =>
        setPanel({
          action: "edit",
          path: props.blockPath,
          block: props.blockProps,
        })
      }
    />
  );
};

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageEditor();

  if (mode === "edit") return <EditModeDialog {...props} />;

  return <DialogBaseComponent {...props} />;
};
