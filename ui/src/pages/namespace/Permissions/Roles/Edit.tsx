import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, Save } from "lucide-react";
import {
  RoleFormSchema,
  RoleFormSchemaType,
  RoleSchemaType,
} from "~/api/enterprise/roles/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import RoleForm from "./Form";
import { useEditRole } from "~/api/enterprise/roles/mutation/edit";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const EditRole = ({
  group,
  close,
  unallowedNames,
}: {
  group: RoleSchemaType;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const { mutate: editGroup, isPending } = useEditRole({
    onSuccess: () => {
      close();
    },
  });

  const resolver = zodResolver(
    RoleFormSchema.refine(
      (x) => !(unallowedNames ?? []).some((n) => n === x.name),
      {
        path: ["group"],
        message: t("pages.permissions.roles.form.name.alreadyExist"),
      }
    )
  );

  const form = useForm<RoleFormSchemaType>({
    defaultValues: {
      name: group.name,
      description: group.description,
      oidcGroups: group.oidcGroups,
      permissions: group.permissions,
    },
    resolver,
  });

  const {
    formState: { isDirty, isValid, isSubmitted },
  } = form;

  const onSubmit: SubmitHandler<RoleFormSchemaType> = (params) => {
    editGroup({
      roleName: group.name,
      payload: params,
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `edit-group`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.permissions.roles.edit.title")}
        </DialogTitle>
      </DialogHeader>
      <RoleForm form={form} onSubmit={onSubmit} formId={formId} />
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.roles.edit.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <Save />}
          {t("pages.permissions.roles.edit.saveBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default EditRole;
