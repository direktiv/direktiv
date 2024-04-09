import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { PlusCircle } from "lucide-react";
import { VariableForm } from "../../../../components/VariableForm";
import { useCreateVar } from "~/api/variables/mutate/create";
import { useTranslation } from "react-i18next";

const defaultMimeType = "application/json";

type CreateProps = { onSuccess: () => void; unallowedNames: string[] };

const Create = ({ onSuccess, unallowedNames }: CreateProps) => {
  const { t } = useTranslation();
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  return (
    <DialogContent>
      <VariableForm
        unallowedNames={unallowedNames}
        defaultValues={{
          name: "",
          data: "",
          mimeType: defaultMimeType,
        }}
        dialogTitle={
          <DialogTitle>
            <PlusCircle />
            {t("pages.settings.variables.create.title")}
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
        onMutate={createVar}
      />
    </DialogContent>
  );
};

export default Create;
