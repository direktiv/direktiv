import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { PlusCircle } from "lucide-react";
import { SecretForm } from "./Form";
import { useCreateSecret } from "~/api/secrets/mutate/create";
import { useTranslation } from "react-i18next";

type CreateProps = { onSuccess: () => void; unallowedNames: string[] };

const Create = ({ onSuccess, unallowedNames }: CreateProps) => {
  const { t } = useTranslation();
  const { mutate: createSecret } = useCreateSecret({
    onSuccess,
  });

  return (
    <DialogContent>
      <SecretForm
        unallowedNames={unallowedNames}
        defaultValues={{
          name: "",
          data: "",
        }}
        dialogTitle={
          <DialogTitle>
            <PlusCircle />
            {t("pages.settings.secrets.create.description")}
          </DialogTitle>
        }
        dialogFooter={
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit">{t("components.button.label.create")}</Button>
          </DialogFooter>
        }
        onMutate={createSecret}
      />
    </DialogContent>
  );
};

export default Create;
