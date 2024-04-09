import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { PlusCircle } from "lucide-react";
import { VarFormCreateEditSchemaType } from "~/api/variables/schema";
import VariableForm from "~/components/VariableForm";
import { useCreateVar } from "~/api/variables/mutate/create";
import { useTranslation } from "react-i18next";

const defaultMimeType = "application/json";

type CreateProps = {
  onSuccess: () => void;
  path: string;
  unallowedNames: string[];
};

const Create = ({ onSuccess, path, unallowedNames }: CreateProps) => {
  const { t } = useTranslation();

  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  const onMutate = (data: VarFormCreateEditSchemaType) => {
    createVar({
      ...data,
      workflowPath: path,
    });
  };

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
            {t("pages.explorer.tree.workflow.settings.variables.create.title")}
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

export default Create;
