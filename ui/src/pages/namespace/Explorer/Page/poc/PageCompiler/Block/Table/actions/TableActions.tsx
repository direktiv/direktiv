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
import { Settings } from "lucide-react";
import { StopPropagation } from "~/components/StopPropagation";
import { TableActionsType } from "../../../../schema/blocks/table";

type TableActionsProps = {
  actions: TableActionsType;
  blockPath: BlockPathType;
};

export const TableActions = ({ actions, blockPath }: TableActionsProps) => (
  <ActionsDialog
    actions={actions}
    blockPath={blockPath}
    renderTrigger={(setOpenedDialogIndex) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <ButtonDesignComponent size="sm" icon>
            <Settings />
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
);
