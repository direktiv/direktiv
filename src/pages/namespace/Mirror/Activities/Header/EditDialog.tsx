import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";

import Button from "~/design/Button";
import { FileCog } from "lucide-react";
import NamespaceCreate from "~/componentsNext/NamespaceCreate";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const EditDialog = () => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const { t } = useTranslation();

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger>
        <Button variant="outline" className="max-md:w-full">
          <FileCog />
          {t("pages.mirror.header.editMirror")}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <NamespaceCreate close={() => setDialogOpen(false)} />
      </DialogContent>
    </Dialog>
  );
};

export default EditDialog;
