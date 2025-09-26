import { Block, BlockPathType } from "../..";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
} from "~/design/Dropdown";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";
import {
  RowActionsType,
  TableActionsType,
} from "../../../../schema/blocks/table";

import { BlockList } from "../../utils/BlockList";
import { StopPropagation } from "~/components/StopPropagation";
import { TemplateString } from "../../../primitives/TemplateString";
import { useState } from "react";

type ActionsDialogProps = {
  actions: RowActionsType | TableActionsType;
  blockPath: BlockPathType;
  renderTrigger: () => React.ReactNode;
};

export const ActionsDialog = ({
  actions,
  blockPath,
  renderTrigger,
}: ActionsDialogProps) => {
  const [openedDialogIndex, setOpenedDialogIndex] = useState<number | null>(
    null
  );

  return (
    <LocalDialog open={openedDialogIndex !== null}>
      <DropdownMenu>
        {renderTrigger()}
        <StopPropagation>
          <DropdownMenuContent align="end">
            {actions.blocks.map((dialog, index) => (
              <DropdownMenuItem key={index} asChild>
                <DialogTrigger
                  onClick={(event) => {
                    event.stopPropagation();
                    setOpenedDialogIndex(index);
                  }}
                  className="text-left"
                >
                  <div>
                    <TemplateString value={dialog.trigger.label} />
                  </div>
                </DialogTrigger>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </StopPropagation>
      </DropdownMenu>
      <div onClick={(event) => event.preventDefault()}>
        <LocalDialogContent>
          <DialogXClose onClick={() => setOpenedDialogIndex(null)} />
          <div className="max-h-[55vh] overflow-y-auto p-2 pt-4">
            {openedDialogIndex !== null && (
              <BlockList path={[...blockPath, openedDialogIndex]}>
                {(actions.blocks[openedDialogIndex]?.blocks ?? []).map(
                  (block, idx) => (
                    <Block
                      key={idx}
                      block={block}
                      blockPath={[...blockPath, openedDialogIndex, idx]}
                    />
                  )
                )}
              </BlockList>
            )}
          </div>
        </LocalDialogContent>
      </div>
    </LocalDialog>
  );
};
