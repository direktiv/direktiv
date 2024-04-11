import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { PlusCircle } from "lucide-react";
import { VarFormCreateEditSchemaType } from "~/api/variables/schema";
import VariableForm from ".";
import { useTranslation } from "react-i18next";

const defaultMimeType = "application/json";

type CreateFormProps = {
  unallowedNames: string[];
  onMutate: (data: VarFormCreateEditSchemaType) => void;
};

export const CreateVariableForm = ({
  unallowedNames,
  onMutate,
}: CreateFormProps) => {
  const { t } = useTranslation();
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
            {t("components.variableForm.title.create")}
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
        onMutate={onMutate}
      />
    </DialogContent>
  );
};
