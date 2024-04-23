import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { RefreshCcw } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";
import { useSync } from "~/api/syncs/mutate/sync";
import { useTranslation } from "react-i18next";

const SyncDialog = ({ loading }: { loading: boolean }) => {
  const [syncModal, setSyncModal] = useState(false);
  const { mutate: performSync } = useSync({
    onSuccess: () => setSyncModal(false),
  });
  const namespace = useNamespace();
  const { t } = useTranslation();

  return (
    <Dialog open={syncModal} onOpenChange={setSyncModal}>
      <DialogTrigger asChild>
        <Button variant="primary" loading={loading} className="max-md:w-full">
          {!loading && <RefreshCcw />}
          {t("pages.mirror.header.sync")}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            <RefreshCcw />
            {t("pages.mirror.syncDialog.title", { namespace: name })}
          </DialogTitle>
        </DialogHeader>
        <p>{t("pages.mirror.syncDialog.description")}</p>
        <Alert variant="warning" className="mb-2">
          {t("pages.mirror.syncDialog.warning")}
        </Alert>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button onClick={() => performSync({ namespace })}>
            {t("pages.mirror.syncDialog.confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default SyncDialog;
