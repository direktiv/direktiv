import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { NodeSchemaType } from "~/api/tree/schema";
import { Trash } from "lucide-react";
import { useDeleteNode } from "~/api/tree/mutate/deleteNode";
import { useTranslation } from "react-i18next";

const Delete = ({
  node,
  close,
}: {
  node: NodeSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteNode, isLoading } = useDeleteNode({
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
      <div
        className="my-3"
        dangerouslySetInnerHTML={{
          __html:
            t("pages.explorer.tree.delete.commonMsg", {
              name: `${node.name}`,
            }) +
            "&nbsp;" +
            (node.type === "directory" &&
              t("pages.explorer.tree.delete.directoryMsg")),
        }}
      />
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.delete.cancel")}
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
          {t("pages.explorer.tree.delete.title")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
