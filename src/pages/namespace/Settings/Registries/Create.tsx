import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  RegistryFormSchema,
  RegistryFormSchemaType,
} from "~/api/registries/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { useCreateRegistry } from "~/api/registries/mutate/createRegistry";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();

  const { mutate: createRegistryMutation } = useCreateRegistry({
    onSuccess,
  });

  const onSubmit: SubmitHandler<RegistryFormSchemaType> = (data) => {
    createRegistryMutation(data);
  };

  const { register, handleSubmit } = useForm<RegistryFormSchemaType>({
    resolver: zodResolver(RegistryFormSchema),
  });

  return (
    <DialogContent>
      <form
        id="create-registry"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            <PlusCircle />
            {t("pages.settings.registries.create.description")}
          </DialogTitle>
        </DialogHeader>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.registries.create.url")}
          </label>
          <Input
            data-testid="new-registry-url"
            placeholder="https://example.com/registry"
            {...register("url")}
          />
        </fieldset>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.registries.create.user")}
          </label>
          <Input
            className="h-96"
            data-testid="new-registry-user"
            placeholder="user-name"
            {...register("user")}
          />
        </fieldset>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.registries.create.password")}
          </label>
          <Input
            className="h-96"
            data-testid="new-registry-pwd"
            type="password"
            placeholder="password"
            {...register("password")}
          />
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button
            data-testid="registry-create-submit"
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
