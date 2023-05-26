import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { VarFormSchema, VarFormSchemaType } from "~/api/vars/schema";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { Textarea } from "~/design/TextArea";
import { useTranslation } from "react-i18next";
import { useUpdateVar } from "~/api/vars/mutate/updateVar";
import { zodResolver } from "@hookform/resolvers/zod";

// TODO: This is almost the same as the Edit component. Consolidate them into one.

type createProps = { onSuccess: () => void };

const Create = ({ onSuccess }: createProps) => {
  const { t } = useTranslation();

  const { register, handleSubmit } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
  });

  const { mutate: createVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    createVarMutation(data);
  };

  return (
    <DialogContent>
      <form
        id="create-variable"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            <PlusCircle />
            {t("pages.settings.variables.create.title")}
          </DialogTitle>
        </DialogHeader>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.variables.create.name")}
          </label>
          <Input
            data-testid="new-variable-name"
            placeholder="variable-name"
            {...register("name")}
          />
        </fieldset>

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
            data-testid="variable-create-submit"
            type="submit"
            variant="primary"
          >
            {t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Create;
