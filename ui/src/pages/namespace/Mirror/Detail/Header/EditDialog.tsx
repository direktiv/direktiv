import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";

import Button from "~/design/Button";
import { FileCog } from "lucide-react";
import { MirrorSchemaType } from "~/api/namespaces/schema/mirror";
import NamespaceEdit from "~/components/NamespaceEdit";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const EditDialog = ({ mirror }: { mirror: MirrorSchemaType }) => {
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
        <NamespaceEdit mirror={mirror} close={() => setDialogOpen(false)} />
      </DialogContent>
    </Dialog>
  );
};

export default EditDialog;
