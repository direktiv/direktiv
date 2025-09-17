import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";

import { Button } from "../Button";
import ButtonDesignComponent from "~/design/Button";
import { MoreVertical } from "lucide-react";
import { TableActionsType } from "../../../schema/blocks/table";
import { TableCell as TableCellDesignComponent } from "~/design/Table";

type ActionsCellProps = {
  actions: TableActionsType;
};

export const ActionsCell = ({ actions }: ActionsCellProps) => {
  const dialogs = actions.blocks.filter((x) => x.type === "dialog");

  return (
    <TableCellDesignComponent className="w-0">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <ButtonDesignComponent variant="ghost" size="sm" icon>
            <MoreVertical />
          </ButtonDesignComponent>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-40" align="end">
          {dialogs.map((dialog, index) => (
            <DropdownMenuItem key={index}>
              <Button
                blockProps={{
                  type: "button",
                  label: dialog.trigger.label,
                }}
                as="text"
              />
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </TableCellDesignComponent>
  );
};
