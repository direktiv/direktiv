import { Block, BlockPathType } from "..";
import { DialogTrigger, DialogXClose } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";

import { BlockList } from "../utils/BlockList";
import ButtonDesignComponent from "~/design/Button";
import { MoreVertical } from "lucide-react";
import { RowActionsType } from "../../../schema/blocks/table";
import { StopPropagation } from "~/components/StopPropagation";
import { TableCell as TableCellDesignComponent } from "~/design/Table";
import { TemplateString } from "../../primitives/TemplateString";
import { useState } from "react";

type RowActionsProps = {
  actions: RowActionsType;
  blockPath: BlockPathType;
};

export const RowActions = ({ actions, blockPath }: RowActionsProps) => {
  const [openedDialogIndex, setOpenedDialogIndex] = useState<number | null>(
    null
  );
  return (
    <TableCellDesignComponent className="w-0">
      <LocalDialog open={openedDialogIndex !== null}>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <ButtonDesignComponent variant="ghost" size="sm" icon>
              <MoreVertical />
            </ButtonDesignComponent>
          </DropdownMenuTrigger>
          <StopPropagation>
            <DropdownMenuContent align="end">
              {actions.blocks.map((dialog, index) => (
                <DropdownMenuItem key={index} asChild>
                  <DialogTrigger
                    onClick={(event) => {
                      event.stopPropagation();
                      setOpenedDialogIndex(index);
                    }}
                    className="w-full text-left"
                  >
                    <TemplateString value={dialog.trigger.label} />
                  </DialogTrigger>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </StopPropagation>
        </DropdownMenu>
        <div onClick={(event) => event.preventDefault()}>
          <LocalDialogContent>
            <DialogXClose
              onClick={() => {
                setOpenedDialogIndex(null);
              }}
            />
            <div className="max-h-[55vh] overflow-y-auto p-2 pt-4">
              {openedDialogIndex !== null && (
                <BlockList path={[...blockPath, 1, openedDialogIndex]}>
                  {(actions.blocks[openedDialogIndex]?.blocks ?? []).map(
                    (block, index) => (
                      <Block
                        key={index}
                        block={block}
                        blockPath={[...blockPath, 1, openedDialogIndex, index]}
                      />
                    )
                  )}
                </BlockList>
              )}
            </div>
          </LocalDialogContent>
        </div>
      </LocalDialog>
    </TableCellDesignComponent>
  );
};
