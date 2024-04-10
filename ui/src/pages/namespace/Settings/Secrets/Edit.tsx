import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";
import {
  SecretFormCreateEditSchemaType,
  SecretSchemaType,
} from "~/api/secrets/schema";

import Button from "~/design/Button";
import { SecretForm } from "./Form";
import { SquareAsterisk } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useUpdateSecret } from "~/api/secrets/mutate/update";

type EditProps = { secret: SecretSchemaType; onSuccess: () => void };

const Edit = ({ onSuccess, secret }: EditProps) => {
  const { t } = useTranslation();
  const { mutate: updateSecret } = useUpdateSecret({
    onSuccess,
  });

  const onMutate = (data: SecretFormCreateEditSchemaType) => {
    updateSecret({ ...data, name: secret.name });
  };

  return (
    <DialogContent>
      <SecretForm
        disableNameInput
        defaultValues={{
          name: secret.name,
          data: "",
        }}
        dialogTitle={
          <DialogTitle>
            <SquareAsterisk />
            {t("pages.settings.secrets.edit.title", {
              name: secret.name,
            })}
          </DialogTitle>
        }
        infoMessage={t("pages.settings.secrets.edit.editNote")}
        dialogFooter={
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit">{t("components.button.label.save")}</Button>
          </DialogFooter>
        }
        onMutate={onMutate}
      />
    </DialogContent>
  );
};

export default Edit;
