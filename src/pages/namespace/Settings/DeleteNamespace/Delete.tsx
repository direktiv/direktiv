import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { Trash } from "lucide-react";
import { useState } from "react";

type DeleteProps = {
  namespace: string;
  onConfirm: () => void;
};

const Delete = ({ namespace, onConfirm }: DeleteProps) => {
  const { t } = useTranslation();

  const [confirmText, setConfirmText] = useState("");
  const [submitDisabled, setSubmitDisabled] = useState(true);

  const onInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const typedValue = e.target.value;
    setConfirmText(typedValue);
    setSubmitDisabled(typedValue !== namespace);
  };

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("components.dialog.header.confirm")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.settings.deleteNamespace.modal.description"
          values={{ namespace }}
        />
        <br />
        <br />

        <div className="flex flex-col gap-2">
          <label htmlFor="confirm">
            <Trans
              i18nKey="pages.settings.deleteNamespace.modal.confirm"
              values={{ namespace }}
            />
          </label>
          <Input id="confirm" value={confirmText} onChange={onInputChange} />
        </div>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button
          data-testid="namespace-delete-confirm"
          onClick={onConfirm}
          variant="destructive"
          disabled={submitDisabled}
        >
          {t("components.button.label.delete")}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

export default Delete;
