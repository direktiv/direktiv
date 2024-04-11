import { DialogClose, DialogFooter, DialogTitle } from "~/design/Dialog";
import {
  VarDetailsSchema,
  VarFormCreateEditSchemaType,
} from "~/api/variables/schema";

import Button from "~/design/Button";
import { FileJson } from "lucide-react";
import VariableForm from ".";
import { useTranslation } from "react-i18next";

type EditVariableProps = {
  variable: VarDetailsSchema;
  unallowedNames: string[];
  onMutate: (data: VarFormCreateEditSchemaType) => void;
};

export const EditVariableForm = ({
  variable,
  unallowedNames,
  onMutate,
}: EditVariableProps) => {
  const { t } = useTranslation();
  const { name, data, mimeType } = variable;
  return (
    <VariableForm
      unallowedNames={unallowedNames}
      defaultValues={{ name, data, mimeType }}
      dialogTitle={
        <DialogTitle>
          <FileJson />
          {t("components.variableForm.title.edit", { name })}
        </DialogTitle>
      }
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
  );
};
