import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FileCog, GitCompare, RefreshCcw } from "lucide-react";

import Button from "~/design/Button";
import { useState } from "react";
import { useSyncMirror } from "~/api/tree/mutate/syncMirror";
import { useTranslation } from "react-i18next";

const Header = ({ name, repo }: { name: string; repo: string }) => {
  const [syncModal, setSyncModal] = useState(false);
  const { mutate: performSync } = useSyncMirror({
    onSuccess: () => setSyncModal(false),
  });
  const { t } = useTranslation();

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <GitCompare className="h-5" /> {name}
          </h3>
          <div className="text-sm">{repo}</div>
        </div>
        <div className="flex grow justify-end gap-4">
          <Button variant="outline" className="max-md:w-full">
            <FileCog />
            {t("pages.mirror.header.editMirror")}
          </Button>
          <Dialog open={syncModal} onOpenChange={setSyncModal}>
            <DialogTrigger asChild>
              <Button variant="primary" className="max-md:w-full">
                <RefreshCcw />
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
              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="ghost">
                    {t("components.button.label.cancel")}
                  </Button>
                </DialogClose>
                <Button onClick={() => performSync()}>
                  {t("pages.mirror.syncDialog.confirm")}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </div>
  );
};

export default Header;
