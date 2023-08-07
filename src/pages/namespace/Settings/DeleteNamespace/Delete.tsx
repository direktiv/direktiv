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
import { useDeleteNamespace } from "~/api/namespaces/mutate/deleteNamespace";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";

type DeleteProps = {
  close: () => void;
};

const Delete = ({ close }: DeleteProps) => {
  const { t } = useTranslation();
  const [confirmText, setConfirmText] = useState("");
  const [submitDisabled, setSubmitDisabled] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const namespace = useNamespace();
  const { mutate: deleteNamespace } = useDeleteNamespace({
    onSuccess: () => {
      setIsLoading(false);
      setSubmitDisabled(true);
      setConfirmText("");
      close();
    },
  });

  if (!namespace) return null;

  const onInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const typedValue = e.target.value;
    setConfirmText(typedValue);
    setSubmitDisabled(typedValue !== namespace);
  };

  const onFormSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsLoading(true);
    deleteNamespace({
      namespace,
    });
  };

  const formId = `new-dir-${namespace}`;

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("components.dialog.header.confirm")}
        </DialogTitle>
      </DialogHeader>

      <form className="my-3" id={formId} onSubmit={onFormSubmit}>
        <Trans
          i18nKey="pages.settings.deleteNamespace.modal.description"
          values={{ namespace }}
        />
        <div className="mt-5 flex flex-col gap-2">
          <label htmlFor="confirm">
            <Trans
              i18nKey="pages.settings.deleteNamespace.modal.confirm"
              values={{ namespace }}
            />
          </label>
          <Input id="confirm" value={confirmText} onChange={onInputChange} />
        </div>
      </form>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button
          data-testid="namespace-delete-confirm"
          type="submit"
          form={formId}
          variant="destructive"
          disabled={submitDisabled}
          loading={isLoading}
        >
          {!isLoading && <Trash />}
          {t("components.button.label.delete")}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

export default Delete;
