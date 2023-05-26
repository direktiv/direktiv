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
} from "~/api/vars/schema";

import Button from "~/design/Button";
import { Textarea } from "~/design/TextArea";
import { Trash } from "lucide-react";
import { useUpdateVar } from "~/api/vars/mutate/updateVar";
import { useVarContent } from "~/api/vars/query/useVarContent";
import { zodResolver } from "@hookform/resolvers/zod";

// TODO: This is almost the same as the Create component. Consolidate them into one.

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
};

const Edit = ({ item, onSuccess }: EditProps) => {
  const { t } = useTranslation();

  const varContent = useVarContent(item.name);

  const { register, handleSubmit } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    values: {
      name: item.name,
      content: varContent.data ?? "",
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

        <fieldset className="flex items-start gap-5">
          <Textarea
            className="h-96"
            data-testid="variable-editor"
            {...register("content")}
          />
        </fieldset>

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
