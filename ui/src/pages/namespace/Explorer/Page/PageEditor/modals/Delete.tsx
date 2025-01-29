import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { LayoutSchemaType } from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import { Trash } from "lucide-react";
import { useTranslation } from "react-i18next";

const DeleteModal = ({
  layout,
  pageElementID,
  close,
  success,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  close: () => void;
  success: (newLayout: LayoutSchemaType) => void;
}) => {
  const { t } = useTranslation();

  let isPending = false;

  const onDelete = (pageElementID: number) => {
    isPending = true;
    const newLayout = [...layout];

    newLayout.splice(Number(pageElementID), 1);

    success(newLayout);
    isPending = false;
    close();
  };

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> Delete this
        </DialogTitle>
      </DialogHeader>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="node-delete-confirm"
          onClick={() => {
            onDelete(pageElementID);
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

export default DeleteModal;
