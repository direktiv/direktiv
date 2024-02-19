import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FileSchemaType, getFilenameFromPath } from "~/api/filesTree/schema";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { useDeletePath } from "~/api/filesTree/mutate/delete";

const Delete = ({
  node,
  close,
}: {
  node: FileSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteNode, isLoading } = useDeletePath({
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
