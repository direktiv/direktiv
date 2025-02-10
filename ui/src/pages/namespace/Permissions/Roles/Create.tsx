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
import { useCreateRole } from "~/api/enterprise/roles/mutation/create";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const CreateGroup = ({
  close,
  unallowedNames,
}: {
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();

  const { mutate: createGroup, isPending } = useCreateRole({
    onSuccess: () => {
      close();
    },
  });

  const resolver = zodResolver(
    RoleFormSchema.refine(
      (x) => !(unallowedNames ?? []).some((n) => n === x.group),
      {
        path: ["group"],
        message: t("pages.permissions.roles.create.role.alreadyExist"),
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
      group: "",
      description: "",
      permissions: [],
    },
    resolver,
  });

  const onSubmit: SubmitHandler<RoleFormSchemaType> = (params) => {
    createGroup(params);
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-group`;

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
            <label className="w-[90px] text-right text-[14px]" htmlFor="group">
              {t("pages.permissions.roles.create.role.label")}
            </label>
            <Input
              id="group"
              placeholder={t("pages.permissions.roles.create.role.placeholder")}
              autoComplete="off"
              {...register("group")}
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
          {/* <PermissionsSelector
            availablePermissions={availablePermissions ?? []}
            permissions={watch("permissions")}
            setPermissions={(permissions) =>
              setValue("permissions", permissions, {
                shouldDirty: true,
                shouldTouch: true,
                shouldValidate: true,
              })
            }
          /> */}
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

export default CreateGroup;
