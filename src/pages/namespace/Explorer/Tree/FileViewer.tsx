import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { Folder } from "lucide-react";
import { NodeSchemaType } from "~/api/tree/schema";
import { useTranslation } from "react-i18next";

const FileViewer = ({
  node,
  close,
}: {
  node: NodeSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Folder /> {t("pages.explorer.tree.fileViewer.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">...</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button>{t("pages.explorer.tree.fileViewer.closeBtn")}</Button>
        </DialogClose>
      </DialogFooter>
    </>
  );
};

export default FileViewer;
