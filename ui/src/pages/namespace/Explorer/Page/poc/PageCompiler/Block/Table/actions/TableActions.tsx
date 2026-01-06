import { ActionsDialog } from "./ActionsDialog";
import { BlockPathType } from "../..";
import Button from "~/design/Button";
import { DropdownMenuTrigger } from "~/design/Dropdown";
import { Settings } from "lucide-react";
import { TableActionsType } from "../../../../schema/blocks/table";

type TableActionsProps = {
  actions: TableActionsType;
  blockPath: BlockPathType;
};

export const TableActions = ({ actions, blockPath }: TableActionsProps) => (
  <ActionsDialog
    actions={actions}
    blockPath={[...blockPath, 0]}
    renderTrigger={() => (
      <DropdownMenuTrigger asChild>
        <Button size="sm" icon>
          <Settings />
        </Button>
      </DropdownMenuTrigger>
    )}
  />
);
