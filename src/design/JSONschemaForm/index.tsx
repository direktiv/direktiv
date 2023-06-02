import {
  BaseInputTemplateProps,
  RJSFSchema,
  SubmitButtonProps,
  WidgetProps,
  getSubmitButtonOptions,
} from "@rjsf/utils";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "../Button";
import Form from "@rjsf/core";
import Input from "../Input";
import React from "react";
import validator from "@rjsf/validator-ajv8";

const FormInput: React.FunctionComponent<WidgetProps> = (props) => {
  const [val, setVal] = React.useState<string>("");
  return (
    <Input
      className="mb-2 mt-1 w-full"
      type={props.options.inputType}
      required={props.required}
      id={props.id}
      value={val}
      onChange={(e) => {
        props.onChange(e.target.value);
        setVal(e.target.value);
      }}
    />
  );
};

const FormSelect: React.FunctionComponent<WidgetProps> = (props) => {
  return (
    <div className="my-4">
      <Select onValueChange={props.onChange}>
        <SelectTrigger>
          <SelectValue placeholder="Select function type" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {props.options.enumOptions?.map((op: any) => {
              return <SelectItem value={op.value}>{op.label}</SelectItem>;
            })}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
};
function SubmitButton(props: SubmitButtonProps) {
  const { uiSchema } = props;
  const { norender } = getSubmitButtonOptions(uiSchema);
  if (norender) {
    return null;
  }
  return (
    <Button type="submit" className="float-right mt-4">
      Submit
    </Button>
  );
}

function BaseInputTemplate(props: BaseInputTemplateProps) {
  return (
    <Input
      className="mb-2 mt-1 w-full"
      type={props.options.inputType}
      required={props.required}
      id={props.id}
      onChange={(e) => {
        props.onChange(e.target.value);
      }}
    />
  );
}

export interface JSONSchemaFormProps {
  schema: RJSFSchema;
}
export const JSONSchemaForm: React.FC<JSONSchemaFormProps> = (props) => {
  const uiSchema = {
    password: {
      "ui:widget": FormInput,
      "ui:options": {
        inputType: "password",
      },
    },
    age: {
      "ui:widget": FormInput,
      "ui:options": {
        inputType: "number",
      },
    },
    type: {
      "ui:widget": FormSelect,
    },
    "ui:submitButtonOptions": {
      norender: false,
    },
  };

  return (
    <Form
      schema={props.schema!}
      templates={{ BaseInputTemplate, ButtonTemplates: { SubmitButton } }}
      validator={validator}
      uiSchema={uiSchema}
    />
  );
};

JSONSchemaForm.displayName = "JSONSchemaForm";
