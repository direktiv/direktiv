import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { PlusCircle, SquareAsterisk } from "lucide-react";
import {
  SecretFormSchema,
  SecretFormSchemaType,
  SecretSchemaType,
} from "~/api/secrets/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { useTranslation } from "react-i18next";
import { useUpdateSecret } from "~/api/secrets/mutate/updateSecret";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { item?: SecretSchemaType; onSuccess: () => void };

const Create = ({ onSuccess, item }: CreateProps) => {
  const { t } = useTranslation();

  const { mutate: createSecretMutation } = useUpdateSecret({
    onSuccess,
  });

  const onSubmit: SubmitHandler<SecretFormSchemaType> = ({ name, value }) => {
    createSecretMutation({
      name,
      value,
    });
  };

  const {
    register,
    handleSubmit,
    setValue,
    setError,
    formState: { errors },
  } = useForm<SecretFormSchemaType>({
    defaultValues: {
      name: item?.name ?? "",
      value: "",
    },
    resolver: zodResolver(SecretFormSchema),
  });

  const editMode = !!item;

  const onFilepickerChange = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;
    try {
      const fileContent = await file.text();
      setValue("value", fileContent);
    } catch (e) {
      setError("value", {
        message: t("pages.settings.secrets.create.fileError"),
      });
    }
  };

  return (
    <DialogContent>
      <form
        id="create-secret"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            {editMode ? <SquareAsterisk /> : <PlusCircle />}
            {editMode
              ? t("pages.settings.secrets.edit.description", {
                  name: item?.name ?? "",
                })
              : t("pages.settings.secrets.create.description")}
          </DialogTitle>
        </DialogHeader>

        <FormErrors errors={errors} className="mb-5" />
        {!editMode && (
          <fieldset className="flex items-center gap-5">
            <label className="w-[150px] text-right" htmlFor="name">
              {t("pages.settings.secrets.create.name")}
            </label>
            <Input
              id="name"
              data-testid="new-secret-name"
              placeholder="secret-name"
              {...register("name")}
            />
          </fieldset>
        )}

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="file-upload">
            {t("pages.settings.secrets.create.file")}
          </label>
          <Input id="file-upload" type="file" onChange={onFilepickerChange} />
        </fieldset>

        <fieldset className="flex items-start gap-5">
          <Textarea
            className="h-96"
            data-testid="new-secret-editor"
            {...register("value")}
          />
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button data-testid="secret-create-submit" type="submit">
            {editMode
              ? t("components.button.label.save")
              : t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Create;
