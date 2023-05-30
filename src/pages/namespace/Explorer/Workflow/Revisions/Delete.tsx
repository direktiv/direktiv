import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { TrimedRevisionSchemaType } from "~/api/tree/schema";
import { useDeleteRevision } from "~/api/tree/mutate/deleteRevision";
import { useDeleteTag } from "~/api/tree/mutate/deleteTag";

const Delete = ({
  path,
  revision,
  isTag,
  close,
}: {
  path: string;
  revision: TrimedRevisionSchemaType;
  isTag: boolean;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteRevision, isLoading: isLoadingRevision } =
    useDeleteRevision({
      onSuccess: () => {
        close();
      },
    });
  const { mutate: deleteTag, isLoading: isLoadingTag } = useDeleteTag({
    onSuccess: () => {
      close();
    },
  });

  const isLoading = isLoadingRevision || isLoadingTag;

  return (
    <>
      <DialogHeader data-testid="dialog-delete-revision">
        <DialogTitle>
          <Trash />
          {isTag
            ? t("pages.explorer.tree.workflow.revisions.delete.titleTag")
            : t("pages.explorer.tree.workflow.revisions.delete.titleRevision")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3 flex flex-col gap-y-5">
        <div>
          <Trans
            i18nKey="pages.explorer.tree.workflow.revisions.delete.description"
            values={{ name: revision.name }}
          />
        </div>
        {!isTag && (
          <div>
            {t("pages.explorer.tree.workflow.revisions.delete.revisionNote")}
          </div>
        )}
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.workflow.revisions.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            if (isTag) {
              deleteTag({
                path,
                tag: revision.name,
              });
            } else {
              deleteRevision({
                path,
                revision: revision.name,
              });
            }
          }}
          variant="destructive"
          loading={isLoading}
          data-testid="dialog-delete-revision-btn-submit"
        >
          {!isLoading && <Trash />}
          {t("pages.explorer.tree.workflow.revisions.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
