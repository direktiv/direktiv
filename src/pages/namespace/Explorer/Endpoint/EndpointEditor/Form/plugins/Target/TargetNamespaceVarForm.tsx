import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetNamespaceVarFormSchema,
  TargetNamespaceVarFormSchemaType,
} from "../../../schema/plugins/target/targetNamespaceVar";

import Button from "~/design/Button";
import { DialogFooter } from "~/design/Dialog";
import Input from "~/design/Input";
import { treatEmptyStringAsUndefined } from "../utils";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<
  TargetNamespaceVarFormSchemaType["configuration"]
>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetNamespaceVarFormSchemaType) => void;
};

export const TargetNamespaceVarForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TargetNamespaceVarFormSchemaType>({
    resolver: zodResolver(TargetNamespaceVarFormSchema),
    defaultValues: {
      type: "target-namespace-var",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm}>
      {errors?.configuration && (
        <FormErrors
          errors={errors?.configuration as errorsType}
          className="mb-5"
        />
      )}

      <div className="my-3 flex flex-col gap-y-5">
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            namespace (optional)
          </label>
          <Input
            {...register("configuration.namespace", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            variable
          </label>
          <Input {...register("configuration.variable")} />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            content type (optional)
          </label>
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            placeholder="application/json"
          />
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
