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
import { DialogTrigger } from "~/design/Dialog";
import { VarSchemaType } from "~/api/variables/schema";
import { useTranslation } from "react-i18next";

type ItemRowProps = {
  item: VarSchemaType;
  onDelete: (item: VarSchemaType) => void;
  onEdit?: () => void;
  onDownload?: () => void;
  children?: React.ReactNode;
};

const ItemRow = ({
  item,
  onDelete,
  onDownload,
  onEdit,
  children,
}: ItemRowProps) => {
  const { t } = useTranslation();

  return (
    <TableRow data-testid="variable-row">
      <TableCell data-testid="item-name">{children}</TableCell>
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

            {onDownload && (
              <div
                role="button"
                className="w-full"
                data-testid="dropdown-actions-download"
                onClick={onDownload}
              >
                <DropdownMenuItem>
                  <DownloadCloud className="mr-2 h-4 w-4" />
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
