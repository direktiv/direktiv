import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";

import Button from "~/design/Button";
import { FileCog } from "lucide-react";
import { MirrorInfoSchemaType } from "~/api/tree/schema/mirror";
import NamespaceCreate from "~/componentsNext/NamespaceEdit";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const EditDialog = ({ mirror }: { mirror: MirrorInfoSchemaType }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const { t } = useTranslation();

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="max-md:w-full">
          <FileCog />
          {t("pages.mirror.header.editMirror")}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <NamespaceCreate mirror={mirror} close={() => setDialogOpen(false)} />
      </DialogContent>
    </Dialog>
  );
};

export default EditDialog;
