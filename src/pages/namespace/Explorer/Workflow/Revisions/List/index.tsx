import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useEffect, useState } from "react";
import { GitMerge, MoreVertical, Tag, Trash, Undo } from "lucide-react";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import CreateTag from "./CreateTag";
import Delete from "./Delete";
import { Link } from "react-router-dom";
import Revert from "./Revert";
import type { TrimedRevisionSchemaType } from "~/api/tree/schema";
import { pages } from "~/util/router/pages";
import { useNodeRevisions } from "~/api/tree/query/revisions";
import { useNodeTags } from "~/api/tree/query/tags";
import { useRouter } from "~/api/tree/query/router";
import { useTranslation } from "react-i18next";

const RevisionsList: FC = () => {
  const { t } = useTranslation();
  const { path, namespace } = pages.explorer.useParams();
  const { data: revisions, isFetched: isFetchedRevisions } = useNodeRevisions({
    path,
  });
  const { data: tags, isFetched: isFetchedTags } = useNodeTags({ path });
  const isFetched = isFetchedRevisions && isFetchedTags;

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
  const [revert, setRevert] = useState<TrimedRevisionSchemaType | undefined>();

  const { data: router } = useRouter({ path });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRev(undefined);
      setDeleteTag(undefined);
      setCreateTag(undefined);
      setRevert(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;
  if (!path) return null;
  // wait for server data to to avoid layout shift
  if (!isFetched) return null;

  return (
    <>
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <GitMerge className="h-5" />
        {t("pages.explorer.tree.workflow.revisions.list.title")}
      </h3>
      <Card>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <Table>
            <TableBody>
              {revisions?.results?.map((rev, i) => {
                const isTag = tags?.results?.some(
                  (tag) => tag.name === rev.name
                );

                const index = router?.routes?.findIndex(
                  (x) => x.ref === rev.name
                );

                const isLatest = rev.name === "latest";
                const Icon = isTag ? Tag : GitMerge;

                return (
                  <TableRow
                    key={i}
                    className="group"
                    data-testid={`revisions-list-${rev.name}`}
                  >
                    <TableCell className="w-0">
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
                    <TableCell className="w-0 justify-start gap-x-3">
                      {index === 0 && (
                        <Badge data-testid="traffic-distribution-primary">
                          {t(
                            "pages.explorer.tree.workflow.revisions.list.distribution",
                            {
                              count: router?.routes?.[0]?.weight,
                            }
                          )}
                        </Badge>
                      )}
                      {index === 1 && (
                        <Badge
                          data-testid="traffic-distribution-secondary"
                          variant="outline"
                        >
                          {t(
                            "pages.explorer.tree.workflow.revisions.list.distribution",
                            {
                              count: router?.routes?.[1]?.weight,
                            }
                          )}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="flex w-auto justify-end gap-x-3">
                      {!isLatest && (
                        <CopyButton
                          value={rev.name}
                          buttonProps={{
                            variant: "outline",
                            className: "hidden group-hover:inline-flex",
                            size: "sm",
                          }}
                        >
                          {(copied) => (copied ? "copied" : "copy")}
                        </CopyButton>
                      )}
                      {!isLatest && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" icon>
                              <MoreVertical />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="w-44">
                            <DropdownMenuLabel>
                              {t(
                                "pages.explorer.tree.workflow.revisions.list.contextMenu.title"
                              )}
                            </DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            <DialogTrigger
                              className="w-full"
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
                                {t(
                                  "pages.explorer.tree.workflow.revisions.list.contextMenu.delete"
                                )}
                              </DropdownMenuItem>
                            </DialogTrigger>
                            <DialogTrigger
                              className="w-full"
                              onClick={() => {
                                setCreateTag(rev);
                              }}
                            >
                              <DropdownMenuItem>
                                <Tag className="mr-2 h-4 w-4" />
                                {t(
                                  "pages.explorer.tree.workflow.revisions.list.contextMenu.tag"
                                )}
                              </DropdownMenuItem>
                            </DialogTrigger>
                            <DialogTrigger
                              className="w-full"
                              onClick={() => {
                                setRevert(rev);
                              }}
                            >
                              <DropdownMenuItem>
                                <Undo className="mr-2 h-4 w-4" />
                                {t(
                                  "pages.explorer.tree.workflow.revisions.list.contextMenu.revert"
                                )}
                              </DropdownMenuItem>
                            </DialogTrigger>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
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
            {revert && (
              <Revert
                path={path}
                revision={revert}
                close={() => {
                  setDialogOpen(false);
                }}
              />
            )}
          </DialogContent>
        </Dialog>
      </Card>
    </>
  );
};

export default RevisionsList;
