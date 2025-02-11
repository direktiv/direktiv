import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { LayoutSchemaType } from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import { Trash } from "lucide-react";

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
  const elementName = layout ? layout[pageElementID]?.name : "element";
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
          <Trash /> Delete
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.explorer.tree.delete.fileMsg"
          values={{ name: `this ${elementName}` }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
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
