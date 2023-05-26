import { Pencil, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { VarSchemaType } from "~/api/vars/schema";

type ItemRowProps = {
  item: VarSchemaType;
  onDelete: (item: VarSchemaType) => void;
  onEdit?: () => void;
};

const ItemRow = ({ item, onDelete, onEdit }: ItemRowProps) => (
  <TableRow>
    <TableCell>{item.name}</TableCell>
    {onEdit && (
      <TableCell className="w-0 px-0">
        <DialogTrigger asChild data-testid="variable-edit" onClick={onEdit}>
          <Button variant="ghost">
            <Pencil />
          </Button>
        </DialogTrigger>
      </TableCell>
    )}
    <TableCell className="w-0 pl-0">
      <DialogTrigger
        asChild
        data-testid="variable-delete"
        onClick={() => onDelete(item)}
      >
        <Button variant="ghost">
          <Trash />
        </Button>
      </DialogTrigger>
    </TableCell>
  </TableRow>
);

export default ItemRow;
