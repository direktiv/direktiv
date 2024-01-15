import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { RotateCcw } from "lucide-react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { useRebuildService } from "~/api/services/mutate/rebuild";

const Rebuild = ({
  service,
  close,
}: {
  service: ServiceSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: rebuildService, isLoading } = useRebuildService({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <RotateCcw /> {t("pages.services.list.rebuild.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.services.list.rebuild.msg"
          values={{ name: service.name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.list.rebuild.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            rebuildService(service.id);
          }}
          variant="destructive"
          loading={isLoading}
        >
          {!isLoading && <RotateCcw />}
          {t("pages.services.list.rebuild.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Rebuild;
