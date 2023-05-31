import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { Trans, useTranslation } from "react-i18next";
import {
  VarFormSchema,
  VarFormSchemaType,
  VarSchemaType,
} from "~/api/variables/schema";

import Button from "~/design/Button";
import Editor from "~/design/Editor";
import { Trash } from "lucide-react";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useUpdateVar } from "~/api/variables/mutate/updateVariable";
import { useVarContent } from "~/api/variables/query/useVariableContent";
import { zodResolver } from "@hookform/resolvers/zod";

// TODO: This is almost the same as the Create component. Consolidate them into one.

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
};

const Edit = ({ item, onSuccess }: EditProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const varContent = useVarContent(item.name);

  const [value, setValue] = useState<string | undefined>(varContent.data);

  const { handleSubmit } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    values: {
      name: item.name,
      content: value ?? "",
    },
  });

  const { mutate: updateVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    updateVarMutation(data);
  };

  return (
    <DialogContent>
      <form
        id="edit-variable"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            <Trash />{" "}
            <Trans
              i18nKey="pages.settings.variables.edit.title"
              values={{ name: item.name }}
            />
          </DialogTitle>
        </DialogHeader>

        <div className="h-[500px]">
          <Editor
            value={varContent.data}
            onChange={(newData) => {
              setValue(newData);
            }}
            theme={theme ?? undefined}
            data-testid="variable-editor"
          />
        </div>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button
            type="submit"
            data-testid="var-edit-submit"
            variant="destructive"
          >
            {t("components.button.label.save")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Edit;
