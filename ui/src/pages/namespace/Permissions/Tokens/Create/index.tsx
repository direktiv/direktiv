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
import DurationHint from "../../components/DurationHint";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import PermissionsSelector from "../../components/PermisionsSelector";
import ShowToken from "./ShowToken";
import { useCreateToken } from "~/api/enterprise/tokens/mutate/create";
import { usePermissionKeys } from "~/api/enterprise/permissions/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const CreateToken = ({ close }: { close: () => void }) => {
  const { t } = useTranslation();
  const { data: availablePermissions } = usePermissionKeys();
  const { mutate: createToken, isPending } = useCreateToken({
    onSuccess: (data) => {
      setCreatedToken(data.token);
    },
  });

  const {
    register,
    setValue,
    handleSubmit,
    watch,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<TokenFormSchemaType>({
    defaultValues: {
      description: "",
      duration: "",
      permissions: [],
    },
    resolver: zodResolver(TokenFormSchema),
  });

  const [createdToken, setCreatedToken] = useState<string>();

  const onSubmit: SubmitHandler<TokenFormSchemaType> = (params) => {
    createToken(params);
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

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
                  className="w-[90px] text-right text-[14px]"
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
                  className="w-[90px] text-right text-[14px]"
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
                      setValue("duration", duration, {
                        shouldDirty: true,
                        shouldTouch: true,
                        shouldValidate: true,
                      });
                    }}
                  />
                </InputWithButton>
              </fieldset>
              <PermissionsSelector
                availablePermissions={availablePermissions ?? []}
                selectedPermissions={watch("permissions")}
                setPermissions={(permissions) =>
                  setValue("permissions", permissions, {
                    shouldDirty: true,
                    shouldTouch: true,
                    shouldValidate: true,
                  })
                }
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
