import {
  Dialog,
  DialogContent,
  DialogTrigger,
} from "../../../../design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../../design/Dropdown";
import { FC, useEffect, useState } from "react";
import {
  Folder,
  FolderUp,
  MoreVertical,
  Play,
  TextCursorInput,
  Trash,
} from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../../../../design/Table";

import Button from "../../../../design/Button";
import { Card } from "../../../../design/Card";
import Delete from "./Delete";
import ExplorerHeader from "./Header";
import { Link } from "react-router-dom";
import NoResult from "./NoResult";
import { NodeSchemaType } from "../../../../api/tree/schema";
import Rename from "./Rename";
import { analyzePath } from "../../../../util/router/utils";
import moment from "moment";
import { pages } from "../../../../util/router/pages";
import { useNamespace } from "../../../../util/store/namespace";
import { useNodeContent } from "../../../../api/tree/query/node";
import { useTranslation } from "react-i18next";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();
  const { data, isSuccess } = useNodeContent({ path });
  const { parent, isRoot } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);

  // we only want to use one dialog component for the whole list,
  // so when the user clicks on the delete button in the list, we
  // set the pointer to that node for the dialog
  const [deleteNode, setDeleteNode] = useState<NodeSchemaType>();
  const [renameNode, setRenameNode] = useState<NodeSchemaType>();
  const { t } = useTranslation();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteNode(undefined);
      setRenameNode(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;

  const results = data?.children?.results ?? [];
  const showTable = !isRoot || results.length > 0;
  const noResults = isSuccess && results.length === 0;

  return (
    <>
      <ExplorerHeader />
      <div className="p-5">
        <Card>
          {showTable && (
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
              <Table>
                <TableBody>
                  {!isRoot && (
                    <TableRow>
                      <TableCell colSpan={2}>
                        <Link
                          to={pages.explorer.createHref({
                            namespace,
                            path: parent?.absolute,
                          })}
                          className="flex items-center space-x-3 hover:underline"
                        >
                          <FolderUp className="h-5" />
                          <span>
                            {t("pages.explorer.tree.list.oneLevelUp")}
                          </span>
                        </Link>
                      </TableCell>
                    </TableRow>
                  )}
                  {results.map((file) => {
                    const Icon =
                      file.expandedType === "workflow" ? Play : Folder;
                    const linkTarget = pages.explorer.createHref({
                      namespace,
                      path: file.path,
                      subpage:
                        file.expandedType === "workflow"
                          ? "workflow"
                          : undefined,
                    });

                    return (
                      <TableRow
                        key={file.name}
                        data-testid={`explorer-item-${file.name}`}
                      >
                        <TableCell>
                          <div className="flex space-x-3">
                            <Icon className="h-5" />
                            <Link
                              to={linkTarget}
                              className="flex-1 hover:underline"
                            >
                              {file.name}
                            </Link>
                            <span className="text-gray-8 dark:text-gray-dark-8">
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
                                {t(
                                  "pages.explorer.tree.list.contextMenu.title"
                                )}
                              </DropdownMenuLabel>
                              <DropdownMenuSeparator />
                              <DialogTrigger
                                className="w-full"
                                data-testid="node-actions-delete"
                                onClick={() => {
                                  setDeleteNode(file);
                                }}
                              >
                                <DropdownMenuItem>
                                  <Trash className="mr-2 h-4 w-4" />
                                  {t(
                                    "pages.explorer.tree.list.contextMenu.delete"
                                  )}
                                </DropdownMenuItem>
                              </DialogTrigger>
                              <DialogTrigger
                                className="w-full"
                                data-testid="node-actions-rename"
                                onClick={() => {
                                  setRenameNode(file);
                                }}
                              >
                                <DropdownMenuItem>
                                  <TextCursorInput className="mr-2 h-4 w-4" />
                                  {t(
                                    "pages.explorer.tree.list.contextMenu.rename"
                                  )}
                                </DropdownMenuItem>
                              </DialogTrigger>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
              <DialogContent>
                {deleteNode && (
                  <Delete
                    node={deleteNode}
                    close={() => {
                      setDialogOpen(false);
                    }}
                  />
                )}
                {renameNode && (
                  <Rename
                    node={renameNode}
                    close={() => {
                      setDialogOpen(false);
                    }}
                    unallowedNames={
                      data?.children?.results.map((x) => x.name) || []
                    }
                  />
                )}
              </DialogContent>
            </Dialog>
          )}
          {noResults && <NoResult />}
        </Card>
      </div>
    </>
  );
};

export default ExplorerPage;
