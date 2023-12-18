import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  RequestConvertFormSchema,
  RequestConvertFormSchemaType,
} from "../../../schema/plugins/inbound/requestConvert";

import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import { DialogFooter } from "~/design/Dialog";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<RequestConvertFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  omit_body: false,
  omit_headers: false,
  omit_consumer: false,
  omit_queries: false,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: RequestConvertFormSchemaType) => void;
};

export const RequestConvertForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const {
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<RequestConvertFormSchemaType>({
    resolver: zodResolver(RequestConvertFormSchema),
    defaultValues: {
      type: "request-convert",
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
            omit headers
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.omit_headers")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_headers", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            omit queries
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.omit_queries")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_queries", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            omit body
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.omit_body")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_body", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            omit consumer
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.omit_consumer")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_consumer", value);
              }
            }}
          />
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
