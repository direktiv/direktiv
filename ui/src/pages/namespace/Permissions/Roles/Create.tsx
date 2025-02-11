import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  RoleFormSchema,
  RoleFormSchemaType,
} from "~/api/enterprise/roles/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import OidcGroupSelector from "../components/OidcGroupSelector";
import { PermissionsArray } from "~/api/enterprise/schema";
import PermissionsSelector from "../components/PermisionsSelector";
import { useCreateRole } from "~/api/enterprise/roles/mutation/create";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const CreateRole = ({
  close,
  unallowedNames,
}: {
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();

  const { mutate: createRole, isPending } = useCreateRole({
    onSuccess: () => {
      close();
    },
  });

  const resolver = zodResolver(
    RoleFormSchema.refine(
      (x) => !(unallowedNames ?? []).some((n) => n === x.name),
      {
        path: ["group"],
        message: t("pages.permissions.roles.create.name.alreadyExist"),
      }
    )
  );

  const {
    register,
    setValue,
    handleSubmit,
    watch,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<RoleFormSchemaType>({
    defaultValues: {
      name: "",
      description: "",
      oidcGroups: [],
      permissions: [],
    },
    resolver,
  });

  const onSubmit: SubmitHandler<RoleFormSchemaType> = (params) => {
    createRole(params);
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-role`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.permissions.roles.create.title")}
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
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.permissions.roles.create.name.label")}
            </label>
            <Input
              id="name"
              placeholder={t("pages.permissions.roles.create.name.placeholder")}
              autoComplete="off"
              {...register("name")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[90px] text-right text-[14px]"
              htmlFor="description"
            >
              {t("pages.permissions.roles.create.description.label")}
            </label>
            <Input
              id="description"
              placeholder={t(
                "pages.permissions.roles.create.description.placeholder"
              )}
              {...register("description")}
            />
          </fieldset>
          <OidcGroupSelector
            oidcGroups={watch("oidcGroups")}
            onChange={(oidcGroups) => {
              setValue("oidcGroups", oidcGroups, {
                shouldDirty: true,
                shouldTouch: true,
                shouldValidate: true,
              });
            }}
          />
          <PermissionsSelector
            permissions={watch("permissions")}
            onChange={(permissions) => {
              const parsedPermissions = PermissionsArray.safeParse(permissions);
              if (parsedPermissions.success) {
                setValue("permissions", parsedPermissions.data, {
                  shouldDirty: true,
                  shouldTouch: true,
                  shouldValidate: true,
                });
              }
            }}
          />
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.roles.create.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.permissions.roles.create.saveBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateRole;
