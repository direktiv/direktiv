import { FC, useState } from "react";

import Alert from "~/design/Alert";
import { AlertCircle } from "lucide-react";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import { useTranslation } from "react-i18next";

const DeleteNamespace: FC = () => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 pb-2 pt-1 font-bold">
          <AlertCircle className="h-5" />
          {t("pages.settings.deleteNamespace.title")}
        </h3>
      </div>

      <Card className="p-5">
        <Alert variant="error">
          <div className="flex flex-col items-center justify-between gap-5 sm:flex-row">
            <div>{t("pages.settings.deleteNamespace.description")}</div>
            <Button
              data-testid="btn-delete-namespace"
              variant="destructive"
              onClick={() => setDialogOpen(true)}
              className="w-full self-center sm:w-min sm:min-w-max"
            >
              {t("pages.settings.deleteNamespace.deleteBtn")}
            </Button>
          </div>
        </Alert>

        <Delete close={() => setDialogOpen(false)} />
      </Card>
    </Dialog>
  );
};

export default DeleteNamespace;
