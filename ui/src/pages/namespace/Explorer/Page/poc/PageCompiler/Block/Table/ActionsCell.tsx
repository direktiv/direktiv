import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";

import { Button } from "../Button";
import ButtonDesignComponent from "~/design/Button";
import { MoreVertical } from "lucide-react";
import { TableCell as TableCellDesignComponent } from "~/design/Table";
import { TableType } from "../../../schema/blocks/table";

type ActionsCellProps = {
  actions: TableType["actions"];
};

export const ActionsCell = ({ actions }: ActionsCellProps) => (
  <TableCellDesignComponent className="w-0">
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <ButtonDesignComponent variant="ghost" size="sm" icon>
          <MoreVertical />
        </ButtonDesignComponent>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-40" align="end">
        {actions.map((action, index) => (
          <DropdownMenuItem key={index}>
            <Button blockProps={action} as="span" />
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  </TableCellDesignComponent>
);
