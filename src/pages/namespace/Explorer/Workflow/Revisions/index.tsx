import {
  Dialog,
  DialogContent,
  DialogTrigger,
} from "../../../../../design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../../../design/Dropdown";
import { FC, useEffect, useState } from "react";
import { GitMerge, MoreVertical, Tag, Trash } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../../../../../design/Table";

import Badge from "../../../../../design/Badge";
import Button from "../../../../../design/Button";
import { Card } from "../../../../../design/Card";
import CopyButton from "../../../../../design/CopyButton";
import CreateTag from "./CreateTag";
import Delete from "./Delete";
import { Link } from "react-router-dom";
import type { TrimedRevisionSchemaType } from "../../../../../api/tree/schema";
import { pages } from "../../../../../util/router/pages";
import { useNodeRevisions } from "../../../../../api/tree/query/revisions";
import { useNodeTags } from "../../../../../api/tree/query/tags";
import { useTranslation } from "react-i18next";

const WorkflowRevisionsPage: FC = () => {
  const { t } = useTranslation();
  const { path, namespace } = pages.explorer.useParams();
  const { data: revisions } = useNodeRevisions({ path });
  const { data: tags } = useNodeTags({ path });

  const [dialogOpen, setDialogOpen] = useState(false);
  // we only want to use one dialog component for the whole list,
  // so when the user clicks on the delete button in the list, we
  // set the pointer to that revision for the dialog
  const [deleteRev, setDeleteRev] = useState<
    TrimedRevisionSchemaType | undefined
  >();
  const [deleteTag, setDeleteTag] = useState<
    TrimedRevisionSchemaType | undefined
  >();
  const [createTag, setCreateTag] = useState<
    TrimedRevisionSchemaType | undefined
  >();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRev(undefined);
      setDeleteTag(undefined);
      setCreateTag(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;
  if (!path) return null;

  return (
    <div className="p-5">
      <Card className="mb-4 flex gap-x-3 p-4">
        {Array.isArray(tags?.results) &&
          tags?.results?.map((x, i) => <Badge key={i}>{x.name}</Badge>)}
      </Card>
      <Card>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <Table>
            <TableBody>
              {revisions?.results?.map((rev, i) => {
                const isTag = tags?.results?.some(
                  (tag) => tag.name === rev.name
                );
                const Icon = isTag ? Tag : GitMerge;
                return (
                  <TableRow key={i} className="group">
                    <TableCell>
                      <div className="flex space-x-3">
                        <Icon aria-hidden="true" className="h-5" />
                        <Link
                          to={pages.explorer.createHref({
                            namespace,
                            path,
                            subpage: "workflow-revisions",
                            revision: rev.name,
                          })}
                        >
                          {rev.name}
                        </Link>
                      </div>
                    </TableCell>
                    <TableCell className="group flex w-auto justify-end gap-x-3">
                      <CopyButton
                        value={rev.name}
                        buttonProps={{
                          variant: "outline",
                          className: "w-24 hidden group-hover:inline-flex",
                          size: "sm",
                        }}
                      >
                        {(copied) => (copied ? "copied" : "copy")}
                      </CopyButton>
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
                          <DropdownMenuLabel>
                            {t(
                              "pages.explorer.tree.workflow.revisions.list.contextMenu.title"
                            )}
                          </DropdownMenuLabel>
                          <DropdownMenuSeparator />
                          <DialogTrigger
                            onClick={() => {
                              if (isTag) {
                                setDeleteTag(rev);
                              } else {
                                setDeleteRev(rev);
                              }
                            }}
                          >
                            <DropdownMenuItem>
                              <Trash className="mr-2 h-4 w-4" />
                              <span>
                                {t(
                                  "pages.explorer.tree.workflow.revisions.list.contextMenu.delete"
                                )}
                              </span>
                            </DropdownMenuItem>
                          </DialogTrigger>
                          <DialogTrigger
                            onClick={() => {
                              setCreateTag(rev);
                            }}
                          >
                            <DropdownMenuItem>
                              <Tag className="mr-2 h-4 w-4" />
                              <span>
                                {t(
                                  "pages.explorer.tree.workflow.revisions.list.contextMenu.tag"
                                )}
                              </span>
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
            {deleteRev && (
              <Delete
                path={path}
                isTag={false}
                revision={deleteRev}
                close={() => {
                  setDialogOpen(false);
                }}
              />
            )}
            {deleteTag && (
              <Delete
                path={path}
                isTag={true}
                revision={deleteTag}
                close={() => {
                  setDialogOpen(false);
                }}
              />
            )}
            {createTag && (
              <CreateTag
                path={path}
                revision={createTag}
                close={() => {
                  setDialogOpen(false);
                }}
                unallowedNames={tags?.results?.map((x) => x.name) ?? []}
              />
            )}
          </DialogContent>
        </Dialog>
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
