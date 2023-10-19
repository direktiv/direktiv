import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { RevisionSchemaType } from "~/api/services/schema/revisions";
import { Trash } from "lucide-react";
import { useDeleteServiceRevision } from "~/api/services/mutate/deleteRevision";

const Delete = ({
  revision,
  service,
  close,
}: {
  revision: RevisionSchemaType;
  service: string;
  close: () => void;
}) => {
  const { t } = useTranslation();

  const { mutate: deleteServiceRevision, isLoading } = useDeleteServiceRevision(
    {
      onSuccess: () => {
        close();
      },
    }
  );

  const revisionParam = revision.revision || revision.rev;
  if (!revisionParam) {
    throw new Error(
      "prop 'revision' or 'rev' must exist when deleting a revision"
    );
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("pages.services.revision.list.delete.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.services.revision.list.delete.msg"
          values={{ name: revision.name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.revision.list.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            deleteServiceRevision({
              service,
              revision: revisionParam,
            });
          }}
          variant="destructive"
          loading={isLoading}
        >
          {!isLoading && <Trash />}
          {t("pages.services.revision.list.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
