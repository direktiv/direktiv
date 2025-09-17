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
import { useState } from "react";

type ActionsCellProps = {
  actions: RowActionsType;
  blockPath: BlockPathType;
};

export const ActionsCell = ({ actions, blockPath }: ActionsCellProps) => {
  // TODO: remove this if schema can enforce that there are only dialogs
  const dialogs = actions.blocks.filter((x) => x.type === "dialog");
  const [dialogContent, setDialogContent] = useState<number | null>(null);
  return (
    <TableCellDesignComponent className="w-0">
      <LocalDialog open={dialogContent !== null}>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <ButtonDesignComponent variant="ghost" size="sm" icon>
              <MoreVertical />
            </ButtonDesignComponent>
          </DropdownMenuTrigger>
          <StopPropagation>
            <DropdownMenuContent align="end">
              {dialogs.map((dialog, index) => (
                <DropdownMenuItem key={index}>
                  <DialogTrigger
                    onClick={(event) => {
                      event.stopPropagation();
                      setDialogContent(index);
                    }}
                    className="w-full text-left"
                  >
                    {dialog.trigger.label}
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
                setDialogContent(null);
              }}
            />
            <div className="max-h-[55vh] overflow-y-auto p-2 pt-4">
              {dialogContent !== null && (
                <BlockList path={[...blockPath, 1, dialogContent]}>
                  {(dialogs[dialogContent]?.blocks ?? []).map(
                    (block, index) => (
                      <Block
                        key={index}
                        block={block}
                        blockPath={[...blockPath, 1, dialogContent, index]}
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
