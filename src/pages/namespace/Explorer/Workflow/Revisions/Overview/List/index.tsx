import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Table, TableBody } from "~/design/Table";

import { Card } from "~/design/Card";
import CreateTag from "./CreateTag";
import Delete from "./Delete";
import { GitMerge } from "lucide-react";
import Revert from "../../components/Revert";
import RevisionTableRow from "./Row";
import type { TrimmedRevisionSchemaType } from "~/api/tree/schema/node";
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
    TrimmedRevisionSchemaType | undefined
  >();
  const [deleteTag, setDeleteTag] = useState<
    TrimmedRevisionSchemaType | undefined
  >();
  const [createTag, setCreateTag] = useState<
    TrimmedRevisionSchemaType | undefined
  >();
  const [revert, setRevert] = useState<TrimmedRevisionSchemaType | undefined>();

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
    <section className="flex flex-col gap-4">
      <h3 className="flex items-center gap-x-2 font-bold">
        <GitMerge className="h-5" />
        {t("pages.explorer.tree.workflow.revisions.overview.list.title")}
      </h3>
      <Card>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <Table>
            <TableBody>
              {revisions?.results?.map((rev, i) => {
                const isTag =
                  tags?.results?.some((tag) => tag.name === rev.name) ?? false;

                const index =
                  router?.routes?.findIndex((x) => x.ref === rev.name) ?? -1;

                return (
                  <RevisionTableRow
                    isPrimaryTraffic={index === 0}
                    isSecondaryTraffic={index === 1}
                    revision={rev}
                    isTag={isTag}
                    key={i}
                    onDeleteRevClicked={setDeleteRev}
                    onDeleteTagCLicked={setDeleteTag}
                    onRevertClicked={setRevert}
                    onCreateTagClicked={setCreateTag}
                  />
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
    </section>
  );
};

export default RevisionsList;
