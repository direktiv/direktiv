import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { TableActionsType, TableType } from "../../../schema/blocks/table";

import { Button } from "../Button";
import ButtonDesignComponent from "~/design/Button";
import { MoreVertical } from "lucide-react";
import { TableCell as TableCellDesignComponent } from "~/design/Table";

type ActionsCellProps = {
  actions: TableActionsType;
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
        TODO:
        {/* {actions.map((action, index) => (
          <DropdownMenuItem key={index}>
            <Button blockProps={action} as="text" />
          </DropdownMenuItem>
        ))} */}
      </DropdownMenuContent>
    </DropdownMenu>
  </TableCellDesignComponent>
);
