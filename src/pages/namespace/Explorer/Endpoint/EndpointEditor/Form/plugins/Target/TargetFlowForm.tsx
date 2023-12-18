import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetFlowFormSchema,
  TargetFlowFormSchemaType,
} from "../../../schema/plugins/target/targetFlow";

import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import { DialogFooter } from "~/design/Dialog";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetFlowFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  async: false,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetFlowFormSchemaType) => void;
};

export const TargetFlowForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const {
    register,
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<TargetFlowFormSchemaType>({
    resolver: zodResolver(TargetFlowFormSchema),
    defaultValues: {
      type: "target-flow",
      configuration: {
        ...predfinedConfig,
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
            asynchronous
          </label>
          <Checkbox
            checked={getValues("configuration.async")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.async", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            content type (optional)
          </label>
          <div>
            <Input
              {...register("configuration.content_type", {
                setValueAs: (value) => (value === "" ? undefined : value),
              })}
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
