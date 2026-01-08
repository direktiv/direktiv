import { Block, BlockPathType } from ".";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import { DialogType } from "../../schema/blocks/dialog";
import { EditModeDialog } from "../../BlockEditor/PageCompiler/EditModeDialog";
import { usePageStateContext } from "../context/pageCompilerContext";

declare const __IS_PAGESAPP__: boolean;

export type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPathType;
  onOpenChange?: (open: boolean) => void;
};

export const DialogBaseComponent = ({
  blockProps,
  blockPath,
  onOpenChange,
}: DialogProps) => {
  const { blocks, trigger } = blockProps;

  return (
    <div onClick={(event) => event.preventDefault()}>
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

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageStateContext();

  if (__IS_PAGESAPP__ || mode === "live") {
    return <DialogBaseComponent {...props} />;
  }

  return <EditModeDialog {...props} />;
};
