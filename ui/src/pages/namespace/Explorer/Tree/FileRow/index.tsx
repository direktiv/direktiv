import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, TextCursorInput, Trash } from "lucide-react";
import { NodeSchemaType, getFilenameFromPath } from "~/api/filesTree/schema";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { ConditionalLink } from "./ConditionalLink";
import { DialogTrigger } from "~/design/Dialog";
import { fileTypeToIcon } from "~/api/tree/utils";
import moment from "moment";
import { useTranslation } from "react-i18next";

const FileRow = ({
  node,
  namespace,
  onRenameClicked,
  onDeleteClicked,
  onPreviewClicked,
}: {
  node: NodeSchemaType;
  namespace: string;
  onRenameClicked: (node: NodeSchemaType) => void;
  onDeleteClicked: (node: NodeSchemaType) => void;
  onPreviewClicked: (node: NodeSchemaType) => void;
}) => {
  const { t } = useTranslation();
  const Icon = fileTypeToIcon(node.type);

  const filename = getFilenameFromPath(node.path);

  return (
    <TableRow data-testid={`explorer-item-${filename}`}>
      <TableCell>
        <div className="flex space-x-3">
          <Icon className="h-5" />
          <ConditionalLink
            node={node}
            namespace={namespace}
            onPreviewClicked={onPreviewClicked}
          >
            {filename}
          </ConditionalLink>
          <span className="text-gray-9 dark:text-gray-dark-9">
            {moment(node.updatedAt).fromNow()}
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
                onDeleteClicked(node);
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
                onRenameClicked(node);
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
