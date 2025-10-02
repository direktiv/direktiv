import { ActionsDialog } from "./ActionsDialog";
import { BlockPathType } from "../..";
import Button from "~/design/Button";
import { DropdownMenuTrigger } from "~/design/Dropdown";
import { MoreVertical } from "lucide-react";
import { RowActionsType } from "../../../../schema/blocks/table";
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
      renderTrigger={() => (
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" icon>
            <MoreVertical />
          </Button>
        </DropdownMenuTrigger>
      )}
    />
  </TableCellDesignComponent>
);
