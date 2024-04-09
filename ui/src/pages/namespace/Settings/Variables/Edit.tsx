import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";
import {
  VarFormCreateEditSchemaType,
  VarSchemaType,
} from "~/api/variables/schema";

import Button from "~/design/Button";
import { FileJson } from "lucide-react";
import { VariableForm } from "../../../../components/VariableForm";
import { useTranslation } from "react-i18next";
import { useUpdateVar } from "~/api/variables/mutate/update";
import { useVarDetails } from "~/api/variables/query/details";

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
  unallowedNames: string[];
};

const Edit = ({ item, onSuccess, unallowedNames }: EditProps) => {
  const { t } = useTranslation();
  const { data, isSuccess } = useVarDetails(item.id);
  const { mutate: updateVar } = useUpdateVar({
    onSuccess,
  });

  const onMutate = (data: VarFormCreateEditSchemaType) => {
    updateVar({
      id: item.id,
      ...data,
    });
  };

  return (
    <DialogContent>
      {isSuccess && (
        <VariableForm
          unallowedNames={unallowedNames}
          defaultValues={{
            name: data.data.name,
            data: data.data.data,
            mimeType: data.data.mimeType,
          }}
          dialogTitle={
            <DialogTitle>
              <FileJson />
              {t("pages.settings.variables.edit.title", {
                name: data.data.name,
              })}
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
      )}
    </DialogContent>
  );
};

export default Edit;
