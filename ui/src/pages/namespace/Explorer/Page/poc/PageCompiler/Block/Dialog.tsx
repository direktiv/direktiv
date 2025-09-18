import { Block, BlockPathType } from ".";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import { DialogType } from "../../schema/blocks/dialog";
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
  const { blocks, trigger } = blockProps;

  return (
    <div onClick={(event) => event.preventDefault}>
      <LocalDialog onOpenChange={onOpenChange}>
        <DialogTrigger asChild onClick={(event) => event.stopPropagation()}>
          <Button blockProps={trigger} />
        </DialogTrigger>
        <LocalDialogContent>
          <DialogXClose />
          <div className="max-h-[55vh] overflow-y-auto p-2 pt-4">
            <BlockList path={blockPath}>
              {blocks.map((block, index) => (
                <Block
                  key={index}
                  block={block}
                  blockPath={[...blockPath, index]}
                />
              ))}
            </BlockList>
          </div>
        </LocalDialogContent>
      </LocalDialog>
    </div>
  );
};

const EditModeDialog = (props: DialogProps) => {
  const { setDialog } = usePageEditorPanel();

  return (
    <DialogBaseComponent
      {...props}
      onOpenChange={(open) => setDialog(open ? props.blockPath : null)}
    />
  );
};

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageEditor();

  if (mode === "edit") return <EditModeDialog {...props} />;

  return <DialogBaseComponent {...props} />;
};
