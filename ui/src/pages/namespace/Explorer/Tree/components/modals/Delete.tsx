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
import { useDeleteMultipleFiles } from "~/api/files/mutate/deleteMultipleFiles";

const Delete = ({
  files,
  close,
  totalFiles,
}: {
  files: BaseFileSchemaType[];
  close: () => void;
  totalFiles?: number;
}) => {
  const { t } = useTranslation();

  const isSingleFile = files.length === 1;
  const isMultipleFiles = files.length > 1;
  const isAllFiles =
    totalFiles && totalFiles > 1 && files.length === totalFiles;
  const file = files[0];

  // Use the hook for single file deletion
  const { mutate: deleteSingle, isPending } = useDeleteFile({
    onSuccess: close,
  });

  // Use Hook for multiple files deletion
  const { mutate: deleteMultiple } = useDeleteMultipleFiles({
    onSuccess: close,
  });

  const handleDelete = () => {
    if (isMultipleFiles) {
      deleteMultiple({ files });
    } else if (isSingleFile && file) {
      deleteSingle({ file });
    }
  };

  const deleteMessage = () => {
    if (isAllFiles) {
      return t("pages.explorer.tree.delete.allFilesMsg");
    } else if (isMultipleFiles) {
      return (
        <Trans
          i18nKey="pages.explorer.tree.delete.multipleFilesMsg"
          values={{ count: files.length }}
        />
      );
    } else if (isSingleFile) {
      if (file?.type === "directory") {
        return (
          <Trans
            i18nKey="pages.explorer.tree.delete.directoryMsg"
            values={{ name: getFilenameFromPath(file?.path || "") }}
          />
        );
      } else {
        return (
          <Trans
            i18nKey="pages.explorer.tree.delete.fileMsg"
            values={{ name: getFilenameFromPath(file?.path || "") }}
          />
        );
      }
    } else {
      return "";
    }
  };

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("pages.explorer.tree.delete.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">{deleteMessage()}</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="node-delete-confirm"
          onClick={handleDelete}
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
