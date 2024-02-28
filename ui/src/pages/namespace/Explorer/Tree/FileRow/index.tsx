import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, TextCursorInput, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";
import { fileTypeToIcon, getFilenameFromPath } from "~/api/files/utils";

import { BaseFileSchemaType } from "~/api/files/schema";
import Button from "~/design/Button";
import { ConditionalLink } from "./ConditionalLink";
import { DialogTrigger } from "~/design/Dialog";
import moment from "moment";
import { useTranslation } from "react-i18next";

const FileRow = ({
  file,
  namespace,
  onRenameClicked,
  onDeleteClicked,
  onPreviewClicked,
}: {
  file: BaseFileSchemaType;
  namespace: string;
  onRenameClicked: (file: BaseFileSchemaType) => void;
  onDeleteClicked: (file: BaseFileSchemaType) => void;
  onPreviewClicked: (file: BaseFileSchemaType) => void;
}) => {
  const { t } = useTranslation();
  const Icon = fileTypeToIcon(file.type);

  const filename = getFilenameFromPath(file.path);

  return (
    <TableRow data-testid={`explorer-item-${filename}`}>
      <TableCell>
        <div className="flex space-x-3">
          <Icon className="h-5" />
          <ConditionalLink
            file={file}
            namespace={namespace}
            onPreviewClicked={onPreviewClicked}
          >
            {filename}
          </ConditionalLink>
          <span className="text-gray-9 dark:text-gray-dark-9">
            {moment(file.updatedAt).fromNow()}
          </span>
        </div>
      </TableCell>
      <TableCell className="w-0">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              data-testid="dropdown-trg-node-actions"
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
              {t("pages.explorer.tree.list.contextMenu.title")}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DialogTrigger
              className="w-full"
              data-testid="node-actions-delete"
              onClick={() => {
                onDeleteClicked(file);
              }}
            >
              <DropdownMenuItem>
                <Trash className="mr-2 h-4 w-4" />
                {t("pages.explorer.tree.list.contextMenu.delete")}
              </DropdownMenuItem>
            </DialogTrigger>
            <DialogTrigger
              className="w-full"
              data-testid="node-actions-rename"
              onClick={() => {
                onRenameClicked(file);
              }}
            >
              <DropdownMenuItem>
                <TextCursorInput className="mr-2 h-4 w-4" />
                {t("pages.explorer.tree.list.contextMenu.rename")}
              </DropdownMenuItem>
            </DialogTrigger>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
};

export default FileRow;
