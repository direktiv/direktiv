import { DownloadCloud, MoreVertical, Pencil, Trash } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import { DialogTrigger } from "~/design/Dialog";
import { useTranslation } from "react-i18next";

type ItemRowProps<TItem> = {
  item: TItem;
  onDelete: (item: TItem) => void;
  onEdit?: () => void;
  onDownload?: () => void;
  onSelect?: (checked: boolean) => void;
  isSelected?: boolean;
  children?: React.ReactNode;
};

const ItemRow = <ItemType,>({
  item,
  onDelete,
  onDownload,
  onEdit,
  onSelect,
  isSelected,
  children,
}: ItemRowProps<ItemType & { name: string }>) => {
  const { t } = useTranslation();

  return (
    <TableRow data-testid="variable-row">
      <TableCell data-testid="item-name" className="flex items-center">
        {onSelect && (
          <Checkbox
            className="mr-3"
            data-testid="variable-checkbox"
            checked={isSelected}
            onCheckedChange={onSelect}
          />
        )}
        {children}
      </TableCell>
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
                  <Pencil className="mr-2 size-4" />
                  {t("pages.settings.generic.contextMenu.edit")}
                </DropdownMenuItem>
              </DialogTrigger>
            )}

            {onDownload && (
              <div
                role="button"
                className="w-full"
                data-testid="dropdown-actions-download"
                onClick={onDownload}
              >
                <DropdownMenuItem>
                  <DownloadCloud className="mr-2 size-4" />
                  {t("pages.settings.generic.contextMenu.download")}
                </DropdownMenuItem>
              </div>
            )}

            <DialogTrigger
              className="w-full"
              data-testid="dropdown-actions-delete"
              onClick={() => onDelete(item)}
            >
              <DropdownMenuItem>
                <Trash className="mr-2 size-4" />
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
