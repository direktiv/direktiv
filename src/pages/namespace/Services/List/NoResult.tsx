import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Diamond, Layers } from "lucide-react";
import { FC, useState } from "react";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

const NoResult: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);

  return (
    <div className="flex flex-col items-center gap-y-5 p-10">
      <div className="flex flex-col items-center justify-center gap-1">
        <Layers />
        <span className="text-center text-sm">
          {t("pages.services.list.empty.title")}
        </span>
      </div>
      <div className="flex flex-col gap-5 sm:flex-row">
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Diamond />
              {t("pages.services.list.empty.create")}
            </Button>
          </DialogTrigger>

          <DialogContent>
            {/* <NewService close={() => setDialogOpen(false)} /> */}
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default NoResult;
