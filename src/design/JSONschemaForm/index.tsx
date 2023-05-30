import {
  BaseInputTemplateProps,
  RJSFSchema,
  SubmitButtonProps,
  WidgetProps,
  getSubmitButtonOptions,
} from "@rjsf/utils";

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

function SubmitButton(props: SubmitButtonProps) {
  const { uiSchema } = props;
  const { norender } = getSubmitButtonOptions(uiSchema);
  if (norender) {
    return null;
  }
  return <Button type="submit">Submit</Button>;
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
    "ui:submitButtonOptions": {
      norender: false,
    },
  };

  return (
    <Form
      schema={props.schema}
      templates={{ BaseInputTemplate, ButtonTemplates: { SubmitButton } }}
      validator={validator}
      uiSchema={uiSchema}
    />
  );
};

JSONSchemaForm.displayName = "JSONSchemaForm";
