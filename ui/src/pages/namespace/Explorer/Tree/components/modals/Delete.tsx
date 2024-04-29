import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import { BaseFileSchemaType } from "~/api/files/schema";
import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { getFilenameFromPath } from "~/api/files/utils";
import { useDeleteFile } from "~/api/files/mutate/deleteFile";

const Delete = ({
  file,
  close,
}: {
  file: BaseFileSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteNode, isPending } = useDeleteFile({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("pages.explorer.tree.delete.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        {file.type === "directory" ? (
          <Trans
            i18nKey="pages.explorer.tree.delete.directoryMsg"
            values={{ name: getFilenameFromPath(file.path) }}
          />
        ) : (
          <Trans
            i18nKey="pages.explorer.tree.delete.fileMsg"
            values={{ name: getFilenameFromPath(file.path) }}
          />
        )}
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="node-delete-confirm"
          onClick={() => {
            deleteNode({ file });
          }}
          variant="destructive"
          loading={isPending}
        >
          {!isPending && <Trash />}
          {t("pages.explorer.tree.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
