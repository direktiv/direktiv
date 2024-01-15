import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Trash } from "lucide-react";

type DeleteProps = {
  name: string;
  onConfirm: () => void;
};

const Delete = ({ name, onConfirm }: DeleteProps) => {
  const { t } = useTranslation();

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("components.dialog.header.confirm")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.settings.registries.delete.description"
          values={{ name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button
          data-testid="registry-delete-confirm"
          onClick={onConfirm}
          variant="destructive"
        >
          {t("components.button.label.delete")}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

export default Delete;
