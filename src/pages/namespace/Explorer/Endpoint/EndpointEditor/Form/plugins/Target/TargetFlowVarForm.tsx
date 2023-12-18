import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetFlowVarFormSchema,
  TargetFlowVarFormSchemaType,
} from "../../../schema/plugins/target/targetFlowVar";

import Button from "~/design/Button";
import { DialogFooter } from "~/design/Dialog";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: Partial<TargetFlowVarFormSchemaType["configuration"]>;
  onSubmit: (data: TargetFlowVarFormSchemaType) => void;
};

export const TargetFlowVarForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TargetFlowVarFormSchemaType>({
    resolver: zodResolver(TargetFlowVarFormSchema),
    defaultValues: {
      type: "target-flow-var",
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
              setValueAs: (value) => (value === "" ? undefined : value),
            })}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            workflow
          </label>
          <Input {...register("configuration.flow")} />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            variable
          </label>
          <Input {...register("configuration.variable")} />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            content type
          </label>
          <div>
            <Input
              {...register("configuration.content_type")}
              placeholder="application/json"
            />
          </div>
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
