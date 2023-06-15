import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SecretFormSchema, SecretFormSchemaType } from "~/api/secrets/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { Textarea } from "~/design/TextArea";
import { useCreateSecret } from "~/api/secrets/mutate/createSecret";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();

  const { mutate: createSecretMutation } = useCreateSecret({
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
    formState: { errors },
  } = useForm<SecretFormSchemaType>({
    resolver: zodResolver(SecretFormSchema),
  });

  return (
    <DialogContent>
      <form
        id="create-secret"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            <PlusCircle /> {t("pages.settings.secrets.create.description")}
          </DialogTitle>
        </DialogHeader>

        <FormErrors errors={errors} />

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.secrets.create.name")}
          </label>
          <Input
            data-testid="new-secret-name"
            id="name"
            placeholder="secret-name"
            {...register("name")}
          />
        </fieldset>

        <fieldset className="flex items-start gap-5">
          <Textarea
            className="h-96"
            data-testid="new-workflow-editor"
            {...register("value")}
          />
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button
            data-testid="secret-create-submit"
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
