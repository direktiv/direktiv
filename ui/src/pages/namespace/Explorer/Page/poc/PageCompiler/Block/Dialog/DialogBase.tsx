import { Block, BlockPathType } from "..";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";

import { BlockList } from "page-blocklist";
import { Button } from "../Button";
import { DialogType } from "../../../schema/blocks/dialog";

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
