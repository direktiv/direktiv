import { CheckCircle2, CircleDashed, PlusCircle, XCircle } from "lucide-react";
import { ComponentProps, FC, useState } from "react";
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
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { useCreateRegistry } from "~/api/registries/mutate/createRegistry";
import { useTestConnection } from "~/api/registries/mutate/testConnection";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

type ConnectionTestSuccessfull = boolean | null; // null = not tested yet
type ButtonVariants = ComponentProps<typeof Button>["variant"];

const useConnectionButtonState = (
  testSuccessful: ConnectionTestSuccessfull
): {
  variant: ButtonVariants;
  icon: FC<React.SVGProps<SVGSVGElement>>;
  label: string;
} => {
  const { t } = useTranslation();
  let variant: ButtonVariants = "outline";
  let icon: FC<React.SVGProps<SVGSVGElement>> = CircleDashed;

  let label = t("pages.settings.registries.create.testConnectionBtn.label");

  if (testSuccessful === true) {
    variant = "primary";
    icon = CheckCircle2;
    label = t("pages.settings.registries.create.testConnectionBtn.success");
  }
  if (testSuccessful === false) {
    variant = "destructive";
    icon = XCircle;
    label = t("pages.settings.registries.create.testConnectionBtn.error");
  }

  return {
    variant,
    icon,
    label,
  };
};

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();
  const { mutate: createRegistryMutation } = useCreateRegistry({ onSuccess });
  const [testSuccessful, setTestSuccessful] =
    useState<ConnectionTestSuccessfull>(null);

  const connectionButtonState = useConnectionButtonState(testSuccessful);

  const { mutate: testConnection, isLoading: testLoading } = useTestConnection({
    onSuccess: () => {
      setTestSuccessful(true);
    },
    onError: () => {
      setTestSuccessful(false);
    },
  });

  const onSubmit: SubmitHandler<RegistryFormSchemaType> = (data) => {
    createRegistryMutation(data);
  };

  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isValid },
  } = useForm<RegistryFormSchemaType>({
    resolver: zodResolver(RegistryFormSchema),
  });

  const onTestConnectionClick = () => {
    testConnection({
      url: getValues("url"),
      username: getValues("user"),
      password: getValues("password"),
    });
  };

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
        <FormErrors errors={errors} className="mb-5" />
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="url">
            {t("pages.settings.registries.create.url")}
          </label>
          <Input
            id="url"
            data-testid="new-registry-url"
            placeholder="https://example.com/registry"
            {...register("url")}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="user">
            {t("pages.settings.registries.create.user")}
          </label>
          <Input
            id="user"
            data-testid="new-registry-user"
            placeholder="user-name"
            {...register("user")}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="password">
            {t("pages.settings.registries.create.password")}
          </label>
          <Input
            id="password"
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
            data-testid="registry-create-test-connection"
            onClick={onTestConnectionClick}
            loading={testLoading}
            disabled={!isValid || testLoading}
            type="button"
            variant={connectionButtonState.variant}
          >
            {!testLoading && <connectionButtonState.icon />}
            {connectionButtonState.label}
          </Button>
          <Button data-testid="registry-create-submit" type="submit">
            {t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Create;
