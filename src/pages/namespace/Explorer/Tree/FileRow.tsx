import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, PropsWithChildren } from "react";
import { MoreVertical, TextCursorInput, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { Link } from "react-router-dom";
import { NodeSchemaType } from "~/api/tree/schema";
import { fileTypeToIcon } from "~/api/tree/utils";
import moment from "moment";
import { pages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

export const LinkWrapper: FC<
  PropsWithChildren & { file: NodeSchemaType; namespace: string }
> = ({ file, namespace, children }) => {
  const isFile = file.expandedType === "file";
  if (isFile) return <a className="flex-1 hover:underline">{children}</a>;
  const linkTarget = pages.explorer.createHref({
    namespace,
    path: file.path,
    subpage: file.expandedType === "workflow" ? "workflow" : undefined,
  });

  return (
    <Link
      data-testid={`explorer-item-link-${file.name}`}
      to={linkTarget}
      className="flex-1 hover:underline"
    >
      {children}
    </Link>
  );
};

const FileRow = ({
  file,
  namespace,
  onRenameClicked,
  onDeleteClicked,
}: {
  file: NodeSchemaType;
  namespace: string;
  onRenameClicked: (file: NodeSchemaType) => void;
  onDeleteClicked: (file: NodeSchemaType) => void;
}) => {
  const { t } = useTranslation();
  const Icon = fileTypeToIcon(file.expandedType);

  return (
    <TableRow data-testid={`explorer-item-${file.name}`}>
      <TableCell>
        <div className="flex space-x-3">
          <Icon className="h-5" />
          <LinkWrapper file={file} namespace={namespace}>
            {file.name}
          </LinkWrapper>
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
