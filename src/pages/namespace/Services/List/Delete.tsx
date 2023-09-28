import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { LucideIcon, Trash } from "lucide-react";

import Button from "~/design/Button";
import { ReactNode } from "react";
import { useDeleteService } from "~/api/services/mutate/deleteService";
import { useTranslation } from "react-i18next";

const Delete = ({
  service,
  workflow,
  version,
  icon: Icon,
  header,
  message,
  close,
}: {
  service: string;
  workflow?: string;
  version?: string;
  icon: LucideIcon;
  header: string;
  message: ReactNode;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteService, isLoading } = useDeleteService({
    workflow,
    version,
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Icon /> {header}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">{message}</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.list.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            deleteService({
              service,
              workflow,
              version,
            });
          }}
          variant="destructive"
          loading={isLoading}
        >
          {!isLoading && <Trash />}
          {t("pages.services.list.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
