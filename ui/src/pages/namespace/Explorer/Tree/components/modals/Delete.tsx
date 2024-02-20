import { BaseFileSchemaType, getFilenameFromPath } from "~/api/files/schema";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { useDeleteFile } from "~/api/files/mutate/deleteFile";

const Delete = ({
  node,
  close,
}: {
  node: BaseFileSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteNode, isLoading } = useDeleteFile({
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
        {node.type === "directory" ? (
          <Trans
            i18nKey="pages.explorer.tree.delete.directoryMsg"
            values={{ name: getFilenameFromPath(node.path) }}
          />
        ) : (
          <Trans
            i18nKey="pages.explorer.tree.delete.fileMsg"
            values={{ name: getFilenameFromPath(node.path) }}
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
            deleteNode({ node });
          }}
          variant="destructive"
          loading={isLoading}
        >
          {!isLoading && <Trash />}
          {t("pages.explorer.tree.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
