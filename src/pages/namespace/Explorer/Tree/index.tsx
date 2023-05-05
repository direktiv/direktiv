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
  FolderOpen,
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
import { NodeSchemaType } from "../../../../api/tree/schema";
import Rename from "./Rename";
import { analyzePath } from "../../../../util/router/utils";
import clsx from "clsx";
import moment from "moment";
import { pages } from "../../../../util/router/pages";
import { useNamespace } from "../../../../util/store/namespace";
import { useNodeContent } from "../../../../api/tree/query/get";

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

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteNode(undefined);
      setRenameNode(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;

  const results = data?.children?.results ?? [];
  const showTable = !isRoot || results.length > 0;
  const isEmpty = isSuccess && results.length === 0;

  return (
    <div>
      <ExplorerHeader />
      <div className="p-5 text-sm">
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
                          <span>..</span>
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
                      <TableRow key={file.name}>
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
                                variant="ghost"
                                size="sm"
                                onClick={(e) => e.preventDefault()}
                                icon
                              >
                                <MoreVertical />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent className="w-40">
                              <DropdownMenuLabel>Edit</DropdownMenuLabel>
                              <DropdownMenuSeparator />
                              <DialogTrigger
                                onClick={() => {
                                  setDeleteNode(file);
                                }}
                              >
                                <DropdownMenuItem>
                                  <Trash className="mr-2 h-4 w-4" />
                                  <span>Delete</span>
                                </DropdownMenuItem>
                              </DialogTrigger>
                              <DialogTrigger
                                onClick={() => {
                                  setRenameNode(file);
                                }}
                              >
                                <DropdownMenuItem>
                                  <TextCursorInput className="mr-2 h-4 w-4" />
                                  <span>Rename</span>
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
          {isEmpty && (
            <div
              className={clsx(
                "flex items-center justify-center gap-2 p-4 text-gray-8 dark:text-gray-dark-8",
                showTable && "border-t border-gray-5 dark:border-gray-dark-5"
              )}
            >
              <FolderOpen />
              <div>this directoy is empty</div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};

export default ExplorerPage;
