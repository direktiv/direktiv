import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import OidcGroupSelector from "./OidcGroupSelector";
import { PermissionsArray } from "~/api/enterprise/schema";
import PermissionsSelector from "../components/PermisionsSelector";
import { RoleFormSchemaType } from "~/api/enterprise/roles/schema";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type RoleFormProps = {
  formId: string;
  form: UseFormReturn<RoleFormSchemaType>;
  onSubmit: (data: RoleFormSchemaType) => void;
};

const RoleForm = ({ form, onSubmit, formId }: RoleFormProps) => {
  const { t } = useTranslation();
  const {
    setValue,
    watch,
    register,
    handleSubmit,
    formState: { errors },
  } = form;

  return (
    <div className="my-3">
      <FormErrors errors={errors} className="mb-5" />
      <form
        id={formId}
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <fieldset className="flex items-center gap-5">
          <label className="w-[120px] text-right text-[14px]" htmlFor="name">
            {t("pages.permissions.roles.form.name.label")}
          </label>
          <Input
            id="name"
            placeholder={t("pages.permissions.roles.form.name.placeholder")}
            autoComplete="off"
            {...register("name")}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label
            className="w-[120px] text-right text-[14px]"
            htmlFor="description"
          >
            {t("pages.permissions.roles.form.description.label")}
          </label>
          <Input
            id="description"
            placeholder={t(
              "pages.permissions.roles.form.description.placeholder"
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
              setValue("permissions", parsedPermissions.data);
            }
          }}
        />
      </form>
    </div>
  );
};

export default RoleForm;
