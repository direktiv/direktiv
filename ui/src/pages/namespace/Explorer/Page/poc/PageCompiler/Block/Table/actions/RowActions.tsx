import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";

import { ActionsDialog } from "./ActionsDialog";
import { BlockPathType } from "../..";
import ButtonDesignComponent from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { MoreVertical } from "lucide-react";
import { RowActionsType } from "../../../../schema/blocks/table";
import { StopPropagation } from "~/components/StopPropagation";
import { TableCell as TableCellDesignComponent } from "~/design/Table";

type RowActionsProps = {
  actions: RowActionsType;
  blockPath: BlockPathType;
};

export const RowActions = ({ actions, blockPath }: RowActionsProps) => (
  <TableCellDesignComponent className="w-0">
    <ActionsDialog
      actions={actions}
      blockPath={[...blockPath, 1]}
      renderTrigger={(setOpenedDialogIndex) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <ButtonDesignComponent variant="ghost" size="sm" icon>
              <MoreVertical />
            </ButtonDesignComponent>
          </DropdownMenuTrigger>
          <StopPropagation>
            <DropdownMenuContent align="end">
              {actions.blocks.map((dialog, index) => (
                <DropdownMenuItem key={index}>
                  <DialogTrigger
                    onClick={(event) => {
                      event.stopPropagation();
                      setOpenedDialogIndex(index);
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
      )}
    />
  </TableCellDesignComponent>
);
