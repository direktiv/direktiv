import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  TokenFormSchema,
  TokenFormSchemaType,
} from "~/api/enterprise/tokens/schema";

import Button from "~/design/Button";
import DurationHint from "./DurationHint";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { PermissionsArray } from "~/api/enterprise/schema";
import PermissionsSelector from "../../components/PermisionsSelector";
import ShowToken from "./ShowToken";
import { useCreateToken } from "~/api/enterprise/tokens/mutate/create";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateTokenProps = {
  close: () => void;
  unallowedNames?: string[];
};

const CreateToken = ({ unallowedNames, close }: CreateTokenProps) => {
  const { t } = useTranslation();
  const { mutate: createToken, isPending } = useCreateToken({
    onSuccess: (token) => {
      setCreatedToken(token);
    },
  });

  const {
    register,
    setValue,
    handleSubmit,
    watch,
    formState: { isDirty, errors },
  } = useForm<TokenFormSchemaType>({
    defaultValues: {
      name: "",
      description: "",
      duration: "",
      permissions: [],
    },
    resolver: zodResolver(
      TokenFormSchema.refine(
        (token) =>
          /**
           * the length of the array could also be restricted in the payload schema,
           * but this would make it impossible to click the "select all" button for
           * "no permissions", cause this would result in an invalid form. But for the
           * user it might be handy to deselect all permissions and e.g. select one
           * single permission afterwards
           */
          token.permissions.length > 0,
        {
          path: ["permissions"],
          message: t(
            "pages.permissions.tokens.create.permissions.noPermissions"
          ),
        }
      ).refine(
        (token) => !(unallowedNames ?? []).some((n) => n === token.name),
        {
          path: ["name"],
          message: t("pages.permissions.tokens.create.name.alreadyExist"),
        }
      )
    ),
  });

  const [createdToken, setCreatedToken] = useState<string>();

  const onSubmit: SubmitHandler<TokenFormSchemaType> = (params) => {
    createToken(params);
  };

  const disableSubmit = !isDirty;

  const formId = `new-token`;

  return (
    <>
      {createdToken ? (
        <ShowToken token={createdToken} onCloseClicked={close} />
      ) : (
        <>
          <DialogHeader>
            <DialogTitle>
              <Diamond /> {t("pages.permissions.tokens.create.title")}
            </DialogTitle>
          </DialogHeader>
          <div className="my-3">
            <FormErrors errors={errors} className="mb-5" />
            <form
              id={formId}
              onSubmit={handleSubmit(onSubmit)}
              className="flex flex-col space-y-5"
            >
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[120px] text-right text-[14px]"
                  htmlFor="name"
                >
                  {t("pages.permissions.tokens.create.name.label")}
                </label>
                <Input
                  id="name"
                  placeholder={t(
                    "pages.permissions.tokens.create.name.placeholder"
                  )}
                  autoComplete="off"
                  {...register("name")}
                />
              </fieldset>
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[120px] text-right text-[14px]"
                  htmlFor="description"
                >
                  {t("pages.permissions.tokens.create.description.label")}
                </label>
                <Input
                  id="description"
                  placeholder={t(
                    "pages.permissions.tokens.create.description.placeholder"
                  )}
                  autoComplete="off"
                  {...register("description")}
                />
              </fieldset>
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[120px] text-right text-[14px]"
                  htmlFor="duration"
                >
                  {t("pages.permissions.tokens.create.duration.label")}
                </label>

                <InputWithButton>
                  <Input
                    id="duration"
                    placeholder={t(
                      "pages.permissions.tokens.create.duration.placeholder"
                    )}
                    {...register("duration")}
                  />
                  <DurationHint
                    onDurationSelect={(duration) => {
                      setValue("duration", duration, { shouldDirty: true });
                    }}
                  />
                </InputWithButton>
              </fieldset>
              <PermissionsSelector
                permissions={watch("permissions")}
                onChange={(permissions) => {
                  const parsedPermissions =
                    PermissionsArray.safeParse(permissions);
                  if (parsedPermissions.success) {
                    setValue("permissions", parsedPermissions.data, {
                      shouldDirty: true,
                      shouldTouch: true,
                    });
                  }
                }}
              />
            </form>
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("pages.permissions.tokens.create.cancelBtn")}
              </Button>
            </DialogClose>
            <Button
              type="submit"
              disabled={disableSubmit}
              loading={isPending}
              form={formId}
            >
              {!isPending && <PlusCircle />}
              {t("pages.permissions.tokens.create.createBtn")}
            </Button>
          </DialogFooter>
        </>
      )}
    </>
  );
};

export default CreateToken;
