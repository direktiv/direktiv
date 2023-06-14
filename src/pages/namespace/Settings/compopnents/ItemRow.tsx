import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Pencil, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { useTranslation } from "react-i18next";

type ItemRowProps<TItem> = {
  item: TItem;
  onDelete: (item: TItem) => void;
  onEdit?: () => void;
};

const ItemRow = <ItemType,>({
  item,
  onDelete,
  onEdit,
}: ItemRowProps<ItemType & { name: string }>) => {
  const { t } = useTranslation();

  return (
    <TableRow>
      <TableCell>{item.name}</TableCell>

      <TableCell className="w-0">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              data-testid={`dropdown-trg-item-${item.name}`}
              variant="ghost"
              size="sm"
              onClick={(e) => e.preventDefault()}
              icon
            >
              <MoreVertical />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-40">
            <DropdownMenuLabel>
              {t("pages.settings.generic.contextMenu.title")}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />

            {onEdit && (
              <DialogTrigger
                className="w-full"
                data-testid="dropdown-actions-edit"
                onClick={onEdit}
              >
                <DropdownMenuItem>
                  <Pencil className="mr-2 h-4 w-4" />
                  {t("pages.settings.generic.contextMenu.edit")}
                </DropdownMenuItem>
              </DialogTrigger>
            )}

            <DialogTrigger
              className="w-full"
              data-testid="dropdown-actions-delete"
              onClick={() => onDelete(item)}
            >
              <DropdownMenuItem>
                <Trash className="mr-2 h-4 w-4" />
                {t("pages.settings.generic.contextMenu.delete")}
              </DropdownMenuItem>
            </DialogTrigger>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
};

export default ItemRow;
